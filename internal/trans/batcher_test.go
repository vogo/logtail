/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package trans

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcherFlushOnThreshold(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var flushes []string

	b := NewBatcher(3, 10*time.Second, func(source string, data []byte) error {
		mu.Lock()
		defer mu.Unlock()

		flushes = append(flushes, string(data))

		return nil
	})

	b.Add("src", []byte("line1"))
	b.Add("src", []byte("line2"))
	b.Add("src", []byte("line3"))

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, flushes, 1)
	assert.Equal(t, "line1\nline2\nline3", flushes[0])
}

func TestBatcherFlushOnTimeout(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var flushes []string

	b := NewBatcher(100, 100*time.Millisecond, func(source string, data []byte) error {
		mu.Lock()
		defer mu.Unlock()

		flushes = append(flushes, string(data))

		return nil
	})
	defer b.Stop()

	b.Add("src", []byte("line1"))
	b.Add("src", []byte("line2"))

	// Wait for timeout flush
	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, flushes, 1)
	assert.Equal(t, "line1\nline2", flushes[0])
}

func TestBatcherFlushOnStop(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var flushes []string

	b := NewBatcher(100, 10*time.Second, func(source string, data []byte) error {
		mu.Lock()
		defer mu.Unlock()

		flushes = append(flushes, string(data))

		return nil
	})

	b.Add("src", []byte("line1"))
	b.Add("src", []byte("line2"))

	b.Stop()

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, flushes, 1)
	assert.Equal(t, "line1\nline2", flushes[0])
}

func TestBatcherUsesFirstSource(t *testing.T) {
	t.Parallel()

	var capturedSource string

	b := NewBatcher(3, 10*time.Second, func(source string, data []byte) error {
		capturedSource = source

		return nil
	})

	b.Add("source-a", []byte("line1"))
	b.Add("source-b", []byte("line2"))
	b.Add("source-c", []byte("line3"))

	assert.Equal(t, "source-a", capturedSource)
}

func TestBatcherConcurrentAdds(t *testing.T) {
	t.Parallel()

	var flushCount atomic.Int32

	b := NewBatcher(10, 10*time.Second, func(source string, data []byte) error {
		flushCount.Add(1)

		return nil
	})

	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			b.Add("src", []byte("data"))
		}(i)
	}

	wg.Wait()
	b.Stop()

	// 100 items with batch size 10 = 10 flushes (plus possibly one from Stop if remainder)
	assert.GreaterOrEqual(t, int(flushCount.Load()), 10)
}

func TestBatcherFlushFuncError(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")

	b := NewBatcher(2, 10*time.Second, func(source string, data []byte) error {
		return errTest
	})

	// Should not panic even when flushFunc returns error
	b.Add("src", []byte("line1"))
	b.Add("src", []byte("line2"))
	b.Stop()
}

func TestBatcherNoDuplicateTimer(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var flushCount int

	b := NewBatcher(5, 50*time.Millisecond, func(source string, data []byte) error {
		mu.Lock()
		defer mu.Unlock()

		flushCount++

		return nil
	})
	defer b.Stop()

	// Add one item, wait for timer flush
	b.Add("src", []byte("line1"))
	time.Sleep(80 * time.Millisecond)

	// Add another item, wait for timer flush
	b.Add("src", []byte("line2"))
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Exactly 2 flushes: one per timer
	assert.Equal(t, 2, flushCount)
}

func TestBatcherStoppedIgnoresAdd(t *testing.T) {
	t.Parallel()

	var flushCount atomic.Int32

	b := NewBatcher(2, 10*time.Second, func(source string, data []byte) error {
		flushCount.Add(1)

		return nil
	})

	b.Stop()

	// Add after stop should be ignored
	b.Add("src", []byte("line1"))
	b.Add("src", []byte("line2"))

	assert.Equal(t, int32(0), flushCount.Load())
}

func TestBatcherDefaultTimeout(t *testing.T) {
	t.Parallel()

	b := NewBatcher(100, 0, func(source string, data []byte) error {
		return nil
	})
	defer b.Stop()

	// Should use default timeout of 1s without panicking
	assert.NotNil(t, b)
}

func TestBatcherDataCopy(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var flushed string

	b := NewBatcher(2, 10*time.Second, func(source string, data []byte) error {
		mu.Lock()
		defer mu.Unlock()

		flushed = string(data)

		return nil
	})

	buf := []byte("original")
	b.Add("src", buf)

	// Modify the original buffer
	copy(buf, "modified")

	b.Add("src", []byte("second"))

	mu.Lock()
	defer mu.Unlock()

	// The batcher should have the original data, not the modified one
	assert.Equal(t, "original\nsecond", flushed)
}
