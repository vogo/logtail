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

package work_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/route"
	"github.com/vogo/logtail/internal/work"
	"github.com/vogo/vogo/vsync/vrun"
)

func TestNewRawWorker(t *testing.T) {
	t.Parallel()

	w := work.NewRawWorker("worker-1", "echo hello", false)
	assert.Equal(t, "worker-1", w.ID)
	assert.NotNil(t, w.Routers)
	assert.Empty(t, w.Routers)
}

func TestWorkerWrite_EmptyData(t *testing.T) {
	t.Parallel()

	w := work.NewRawWorker("w1", "echo", false)
	n, err := w.Write(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestWorkerWrite_SingleLine(t *testing.T) {
	t.Parallel()

	runner := vrun.New()

	router := &route.Router{
		Lock:    sync.Mutex{},
		Runner:  runner.NewChild(),
		ID:      "test-router",
		Name:    "test-router",
		Channel: make(chan []byte, 16),
	}

	w := work.NewRawWorker("w1", "echo", false)
	w.Runner = runner
	w.Routers["test-router"] = router

	n, err := w.Write([]byte("hello\n"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	// Read directly from channel — data should already be buffered
	select {
	case data := <-router.Channel:
		assert.Equal(t, "hello\n", string(data))
	default:
		t.Fatal("expected data in router channel")
	}

	router.Stop()
}

func TestWorkerStopRouters(t *testing.T) {
	t.Parallel()

	runner := vrun.New()

	router := &route.Router{
		Lock:    sync.Mutex{},
		Runner:  runner.NewChild(),
		ID:      "r1",
		Name:    "r1",
		Channel: make(chan []byte, 1),
	}

	w := work.NewRawWorker("w1", "echo", false)
	w.Runner = runner
	w.Routers["r1"] = router

	w.StopRouters() // should not panic
}

func TestWorkerNotifyError(t *testing.T) {
	t.Parallel()

	w := work.NewRawWorker("w1", "echo", false)
	w.ErrorChan = make(chan error, 1)

	w.NotifyError(work.ErrWorkerCommandStopped)

	err := <-w.ErrorChan
	assert.ErrorIs(t, err, work.ErrWorkerCommandStopped)
}

func TestWorkerNotifyError_ClosedChannel(t *testing.T) {
	t.Parallel()

	w := work.NewRawWorker("w1", "echo", false)
	w.ErrorChan = make(chan error, 1)

	close(w.ErrorChan)

	// Should not panic (recovers from send on closed channel)
	w.NotifyError(work.ErrWorkerCommandStopped)
}
