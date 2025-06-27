// Copyright 2025 TimeWtr
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package weight

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/TimeWtr/slab/common"
	"github.com/TimeWtr/slab/utils/atomicx"
	"github.com/TimeWtr/slab/utils/log"
	"github.com/fsnotify/fsnotify"
)

const (
	RunningState = iota
	StoppedState
)

type Provider interface {
	Watch() (chan<- common.Config, error)
	Close()
}

type FileProvider struct {
	parseType        ParseType
	filepath         string
	dir              string
	watcher          *fsnotify.Watcher
	ch               chan common.Config
	closeCh          chan struct{}
	state            *atomicx.Int32
	logger           log.Logger
	lock             sync.Mutex
	debounceLock     sync.Mutex
	debounceTimer    *time.Timer
	debounceDuration time.Duration
	debouncePending  *atomicx.Bool
	wg               sync.WaitGroup
}

func NewFileProvider(parseType ParseType, filepath string, logger log.Logger) (*FileProvider, error) {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}

		return nil, fmt.Errorf("failed to stat file %s: %w", filepath, err)
	}

	const debounceTimeout = time.Millisecond * 500
	return &FileProvider{
		parseType:        parseType,
		filepath:         filepath,
		dir:              path.Dir(filepath),
		logger:           logger,
		state:            atomicx.NewInt32(0),
		debounceDuration: debounceTimeout,
		debounceTimer:    time.NewTimer(debounceTimeout),
		debouncePending:  atomicx.NewBool(),
		closeCh:          make(chan struct{}),
	}, nil
}

func (f *FileProvider) Watch() (chan<- common.Config, error) {
	initialCfg, err := f.reloadFile(false)
	if err != nil {
		return nil, err
	}

	f.ch = make(chan common.Config, 100)
	f.ch <- initialCfg

	// create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	f.watcher = watcher

	// add file to watcher
	if err1 := f.watcher.Add(f.dir); err1 != nil {
		return nil, err1
	}

	f.logger.Info("Adding a configuration file to the file watcher",
		log.StringField("file path", f.filepath),
		log.StringField("parse type", f.parseType.String()))

	f.wg.Add(1)
	go f.watchLoop()

	return f.ch, nil
}

func (f *FileProvider) watchLoop() {
	defer func() {
		f.wg.Done()
		f.debounceTimer.Stop()
		if f.watcher != nil {
			if err := f.watcher.Close(); err != nil {
				f.logger.Error("failed to close the file watcher", log.ErrorField(err))
			}
		}

		if r := recover(); r != nil {
			f.logger.Error("file provider error", log.Field{Key: "cause", Val: r})
		}
	}()

	f.logger.Debug("Start monitoring the configuration file",
		log.StringField("file path", f.filepath))

	for {
		select {
		case e, ok := <-f.watcher.Events:
			if !ok {
				return
			}

			if filepath.Clean(e.Name) != filepath.Clean(f.filepath) {
				continue
			}

			f.logger.Debug("File change event detected",
				log.StringField("event", e.Op.String()),
				log.StringField("path", e.Name))

			switch e.Op {
			case fsnotify.Write:
				f.scheduleReload()
			case fsnotify.Create:
				f.logger.Info("New profile creation detected")
				f.scheduleReload()
			case fsnotify.Remove, fsnotify.Rename:
				f.handleFileRemoval()
			case fsnotify.Chmod:
				f.logger.Debug("File changes detected, skipping processing")
			default:
				f.logger.Error("Unhandled file event types",
					log.StringField("event operation", e.Op.String()))
			}

		case <-f.closeCh:
			f.logger.Debug("Received a shutdown signal and exited file monitoring")
			return
		case err, ok := <-f.watcher.Errors:
			if !ok {
				return
			}

			f.logger.Error("File listener error", log.ErrorField(err))
		}
	}
}

func (f *FileProvider) scheduleReload() {
	f.debounceLock.Lock()
	defer f.debounceLock.Unlock()

	if f.debounceTimer != nil {
		f.debounceTimer.Stop()
	}

	if f.state.Load() == StoppedState {
		return
	}

	f.debounceTimer = time.AfterFunc(f.debounceDuration, func() {
		f.debounceLock.Lock()
		defer f.debounceLock.Unlock()

		f.debouncePending.Store(false)
		cfg, err := f.reloadFile(true)
		if err != nil {
			f.logger.Error("Failed to reload config file",
				log.ErrorField(err),
				log.StringField("file", f.filepath),
			)
			return
		}

		f.ch <- cfg
	})

	f.debouncePending.Store(true)
}

func (f *FileProvider) handleFileRemoval() {
	f.logger.Warn("The config file was removed or renamed",
		log.StringField("file", f.filepath))

	const (
		maxAttempts = 5
		retryDelay  = 200 * time.Millisecond
	)

	for i := 0; i < maxAttempts; i++ {
		if _, err := os.Stat(f.filepath); err == nil {
			f.logger.Debug("The configuration file has been restored, rebuild the monitoring")
			if err1 := f.watcher.Add(f.dir); err1 != nil {
				f.logger.Error("Failed to rebuild the monitoring", log.ErrorField(err1))
				continue
			}

			f.scheduleReload()
			return
		}
		time.Sleep(retryDelay)
	}
	f.logger.Error("The configuration file was not restored within the timeout window.")
}

func (f *FileProvider) reloadFile(reload bool) (common.Config, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	bs, err := os.ReadFile(f.filepath)
	if err != nil {
		return common.Config{}, err
	}

	var cfg common.Config
	switch f.parseType {
	case ParseTypeYAML:
		cfg, err = parseYaml(bs)
	case ParseTypeJSON:
		cfg, err = parseJSON(bs)
	case ParseTypeTOML:
		cfg, err = parseToml(bs)
	default:
		return common.Config{}, errors.New("parse type error")
	}

	if err != nil {
		return common.Config{}, err
	}

	if reload {
		f.logger.Info("reload file success")
	}

	return cfg, nil
}

func (f *FileProvider) Close() {
	if !f.state.CompareAndSwap(RunningState, StoppedState) {
		return
	}

	close(f.closeCh)
	f.wg.Wait()
	close(f.ch)
}
