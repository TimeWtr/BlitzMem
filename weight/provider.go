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
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/TimeWtr/slab/common"
	"github.com/TimeWtr/slab/utils/log"
	"github.com/fsnotify/fsnotify"
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
	logger           log.Logger
	lock             sync.Mutex
	debounceLock     sync.Mutex
	debounceTimer    *time.Timer
	debounceDuration time.Duration
	debouncePending  bool
	wg               sync.WaitGroup
}

func NewFileProvider(parseType ParseType, filepath string, logger log.Logger) (*FileProvider, error) {
	dir := path.Dir(filepath)
	_, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	const debounceTimeout = time.Millisecond * 500
	return &FileProvider{
		parseType:        parseType,
		filepath:         filepath,
		dir:              dir,
		logger:           logger,
		debounceDuration: debounceTimeout,
	}, nil
}

func (f *FileProvider) Watch() (chan<- common.Config, error) {
	f.ch = make(chan common.Config, 100)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(f.dir); err != nil {
		return nil, err
	}

	f.wg.Add(1)
	go f.watchLoop()

	return f.ch, nil
}

func (f *FileProvider) watchLoop() {
	defer func() {
		if r := recover(); r != nil {
			f.logger.Error("file provider error", log.Field{
				Key: "cause",
				Val: r,
			})
		}
	}()
	defer f.wg.Done()
	defer f.debounceTimer.Stop()
	defer func() {
		_ = f.watcher.Close()
	}()

	for {
		select {
		case e, ok := <-f.watcher.Events:
			if !ok {
				return
			}

			if filepath.Clean(e.Name) != filepath.Clean(f.filepath) {
				continue
			}

		case <-f.closeCh:
			return
		case <-f.watcher.Errors:
		}
	}
}

func (f *FileProvider) scheduleReload() {
	f.debounceLock.Lock()
	defer f.debounceLock.Unlock()

	if f.debounceTimer != nil {
		f.debounceTimer.Stop()
	}

	f.debounceTimer = time.AfterFunc(f.debounceDuration, func() {
		f.debounceLock.Lock()
		defer f.debounceLock.Unlock()

		f.debouncePending = false
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

	f.debouncePending = true
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
	close(f.closeCh)
	f.wg.Wait()
}
