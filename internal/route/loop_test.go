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

package route_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/gorun"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/route"
	"github.com/vogo/logtail/internal/trans"
)

func TestStartLoop_RoutesData(t *testing.T) {
	t.Parallel()

	var received []byte

	mu := &sync.Mutex{}
	mockTransfer := &mockTransfer{
		transFn: func(source string, data ...[]byte) error {
			mu.Lock()
			defer mu.Unlock()

			for _, d := range data {
				received = append(received, d...)
			}

			return nil
		},
	}

	runner := gorun.New()

	router := &route.Router{
		Lock:      sync.Mutex{},
		Runner:    runner.NewChild(),
		ID:        "loop-test",
		Name:      "loop-test",
		Channel:   make(chan []byte, 16),
		Transfers: []trans.Transfer{mockTransfer},
	}

	go router.StartLoop()

	router.Channel <- []byte("hello world")

	time.Sleep(50 * time.Millisecond)

	router.Stop()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, "hello world", string(received))
	mu.Unlock()
}

func TestStartLoop_StopsOnNilData(t *testing.T) {
	t.Parallel()

	runner := gorun.New()

	router := &route.Router{
		Lock:    sync.Mutex{},
		Runner:  runner.NewChild(),
		ID:      "nil-test",
		Name:    "nil-test",
		Channel: make(chan []byte, 16),
	}

	done := make(chan struct{})

	go func() {
		router.StartLoop()
		close(done)
	}()

	router.Channel <- nil

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("StartLoop did not stop on nil data")
	}
}

func TestRouteWithMatchers(t *testing.T) {
	t.Parallel()

	var received []byte

	mu := &sync.Mutex{}
	mockTrans := &mockTransfer{
		transFn: func(source string, data ...[]byte) error {
			mu.Lock()
			defer mu.Unlock()

			for _, d := range data {
				received = append(received, d...)
			}

			return nil
		},
	}

	router := &route.Router{
		Lock: sync.Mutex{},
		Matchers: []match.Matcher{
			match.NewContainsMatcher("ERROR", true),
		},
		Transfers: []trans.Transfer{mockTrans},
	}

	_ = router.Route([]byte("ERROR something bad"))
	_ = router.Route([]byte("INFO all good"))

	mu.Lock()
	assert.Equal(t, "ERROR something bad", string(received))
	mu.Unlock()
}

func TestRouteNoMatchers(t *testing.T) {
	t.Parallel()

	callCount := 0

	mockTrans := &mockTransfer{
		transFn: func(source string, data ...[]byte) error {
			callCount++
			return nil
		},
	}

	router := &route.Router{
		Lock:      sync.Mutex{},
		Transfers: []trans.Transfer{mockTrans},
	}

	_ = router.Route([]byte("anything"))
	assert.Equal(t, 1, callCount)
}

func TestTransNoTransfers(t *testing.T) {
	t.Parallel()

	router := &route.Router{
		Lock: sync.Mutex{},
	}

	err := router.Trans([]byte("data"))
	assert.NoError(t, err)
}

func TestSetMatchers(t *testing.T) {
	t.Parallel()

	router := &route.Router{Lock: sync.Mutex{}}
	assert.Nil(t, router.Matchers)

	matchers := []match.Matcher{match.NewContainsMatcher("X", true)}
	router.SetMatchers(matchers)
	assert.Len(t, router.Matchers, 1)
}

func TestMatches(t *testing.T) {
	t.Parallel()

	router := &route.Router{
		Lock: sync.Mutex{},
		Matchers: []match.Matcher{
			match.NewContainsMatcher("ERROR", true),
			match.NewContainsMatcher("IGNORE", false),
		},
	}

	assert.True(t, router.Matches([]byte("ERROR real")))
	assert.False(t, router.Matches([]byte("ERROR IGNORE skip")))
	assert.False(t, router.Matches([]byte("INFO no match")))
}

type mockTransfer struct {
	transFn func(source string, data ...[]byte) error
}

func (m *mockTransfer) Name() string { return "mock" }

func (m *mockTransfer) Trans(source string, data ...[]byte) error {
	if m.transFn != nil {
		return m.transFn(source, data...)
	}

	return nil
}

func (m *mockTransfer) Start() error { return nil }
func (m *mockTransfer) Stop() error  { return nil }
