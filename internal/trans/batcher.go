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
	"bytes"
	"sync"
	"time"

	"github.com/vogo/vogo/vlog"
)

const defaultBatchTimeout = time.Second

type batchEntry struct {
	source string
	data   []byte
}

// Batcher buffers log lines and flushes them by count threshold or time window.
type Batcher struct {
	mu           sync.Mutex
	buffer       []batchEntry
	batchSize    int
	batchTimeout time.Duration
	timer        *time.Timer // nil when no timer is active
	flushFunc    func(source string, data []byte) error
	stopped      bool
}

// NewBatcher creates a new Batcher.
func NewBatcher(batchSize int, batchTimeout time.Duration, flushFunc func(string, []byte) error) *Batcher {
	if batchTimeout <= 0 {
		batchTimeout = defaultBatchTimeout
	}

	return &Batcher{
		buffer:       make([]batchEntry, 0, batchSize),
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		flushFunc:    flushFunc,
	}
}

// Add appends data to the batch buffer. It flushes when the batch size threshold is reached.
func (b *Batcher) Add(source string, data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		return
	}

	// Copy data since the caller's buffer may be reused.
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	b.buffer = append(b.buffer, batchEntry{source: source, data: dataCopy})

	if len(b.buffer) >= b.batchSize {
		b.flush()
	}

	// Start a timer if none is active.
	if b.timer == nil && len(b.buffer) > 0 {
		b.timer = time.AfterFunc(b.batchTimeout, func() {
			b.mu.Lock()
			defer b.mu.Unlock()

			if !b.stopped && len(b.buffer) > 0 {
				b.flush()
			}
		})
	}
}

// flush sends buffered entries to flushFunc. MUST be called with b.mu held.
func (b *Batcher) flush() {
	if len(b.buffer) == 0 {
		return
	}

	source := b.buffer[0].source

	parts := make([][]byte, len(b.buffer))
	for i, entry := range b.buffer {
		parts[i] = entry.data
	}

	payload := bytes.Join(parts, []byte("\n"))

	if err := b.flushFunc(source, payload); err != nil {
		vlog.Warnf("batcher flush error: %v", err)
	}

	b.buffer = b.buffer[:0]

	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
}

// Stop flushes remaining data and marks the batcher as stopped.
func (b *Batcher) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.flush()
	b.stopped = true
}
