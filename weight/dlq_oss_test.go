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
	"context"
	"testing"
	"time"

	"github.com/TimeWtr/TurboAlloc/common"
	"github.com/stretchr/testify/assert"
)

var mockEvent = Event{
	category:  common.SmallSizeCategory,
	eventType: SizeClassConfigChange,
	timestamp: time.Now().UnixMilli(),
}

func TestDLQOss_BasicPushPop(t *testing.T) {
	dlq := newDLQOss(32)
	defer dlq.Close()

	err := dlq.Push(context.Background(), &DLQEvent{OriginalEvent: mockEvent})
	assert.NoError(t, err)

	event, err := dlq.Pop(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, mockEvent, event.OriginalEvent)
}

//func TestDLQOss_ConcurrentPushPop(t *testing.T) {
//	dlq := newDLQOss(100)
//	defer dlq.Close()
//
//	const numEvents = 1000
//	var wg sync.WaitGroup
//	received := 0
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for i := 0; i < numEvents; i++ {
//			err := dlq.Push(context.Background(), &DLQEvent{OriginalEvent: mockEvent})
//			if err != nil {
//				t.Errorf("Push error: %v", err)
//				return
//			}
//		}
//	}()
//
//	for i := 0; i < 4; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			for {
//				event, err := dlq.Pop(context.Background())
//				if err != nil {
//					t.Logf("Pop error: %v", err)
//					return
//				}
//				if event == nil {
//					time.Sleep(10 * time.Millisecond)
//					continue
//				}
//
//				received++
//				if received == numEvents {
//					return
//				}
//			}
//		}()
//	}
//
//	wg.Wait()
//
//	assert.Equal(t, numEvents, received)
//}
//
//func TestDLQOss_Expansion(t *testing.T) {
//	dlq := newDLQOss(2)
//	defer dlq.Close()
//
//	for i := 0; i < 50; i++ {
//		err := dlq.Push(context.Background(), &DLQEvent{OriginalEvent: mockEvent})
//		assert.NoError(t, err)
//	}
//
//	err := dlq.Push(context.Background(), &DLQEvent{OriginalEvent: mockEvent})
//	assert.NoError(t, err)
//}
//
//func TestDLQOss_Shrink(t *testing.T) {
//	dlq := newDLQOss(100)
//	defer dlq.Close()
//
//	for i := 0; i < 50; i++ {
//		err := dlq.Push(context.Background(), &DLQEvent{OriginalEvent: mockEvent})
//		assert.NoError(t, err)
//	}
//
//	for i := 0; i < 45; i++ {
//		_, err := dlq.Pop(context.Background())
//		assert.NoError(t, err)
//	}
//
//	for i := 0; i < 10; i++ {
//		_, err := dlq.Pop(context.Background())
//		assert.NoError(t, err)
//	}
//
//	remaining := 0
//	for {
//		event, err := dlq.Pop(context.Background())
//		assert.NoError(t, err)
//		if event == nil {
//			break
//		}
//		remaining++
//	}
//
//	assert.Equal(t, 5, remaining)
//}

func TestDLQOss_Close(t *testing.T) {
	dlq := newDLQOss(100)

	for i := 0; i < 10; i++ {
		err := dlq.Push(context.Background(), &DLQEvent{OriginalEvent: mockEvent})
		assert.NoError(t, err)
	}
	dlq.Close()

	err := dlq.Push(context.Background(), &DLQEvent{OriginalEvent: mockEvent})
	assert.Equal(t, ErrQueueClosed, err)
	event, err := dlq.Pop(context.Background())
	assert.Equal(t, ErrQueueClosed, err)
	assert.Nil(t, event)
}

func TestDLQOss_ContextCancellation(t *testing.T) {
	dlq := newDLQOss(100)
	defer dlq.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := dlq.Push(ctx, &DLQEvent{OriginalEvent: mockEvent})
	assert.Equal(t, context.Canceled, err)

	_, err = dlq.Pop(ctx)
	assert.Equal(t, context.Canceled, err)
}
