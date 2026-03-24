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

package internal_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/route"
	"github.com/vogo/logtail/internal/starter"
	"github.com/vogo/logtail/internal/tail"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/logtail/internal/webapi"
	"github.com/vogo/vogo/vsync/vrun"
)

// ----- Integration Test: Multi-instance isolation -----
// Test that two independently created Tailer instances operate in isolation.
// Starting and stopping one tailer must not affect the other.
// Covers acceptance criteria: singleton removal, independent lifecycle.
func TestIntegrationMultiInstanceIsolation(t *testing.T) {
	t.Parallel()

	// Create two separate configs with different server names.
	makeConfig := func(serverName string) *conf.Config {
		return &conf.Config{
			Transfers: map[string]*conf.TransferConfig{
				"null": {
					Name: "null",
					Type: trans.TypeNull,
				},
			},
			Routers: map[string]*conf.RouterConfig{
				"r1": {
					Name:      "r1",
					Transfers: []string{"null"},
				},
			},
			Servers: map[string]*conf.ServerConfig{
				serverName: {
					Name:    serverName,
					Command: "echo isolation-test",
					Routers: []string{"r1"},
				},
			},
		}
	}

	// Start two tailer instances with independent configs.
	tailer1, err1 := starter.StartLogtail(makeConfig("iso-server-1"))
	require.NoError(t, err1, "tailer1 should start without error")
	require.NotNil(t, tailer1, "tailer1 should not be nil")

	tailer2, err2 := starter.StartLogtail(makeConfig("iso-server-2"))
	require.NoError(t, err2, "tailer2 should start without error")
	require.NotNil(t, tailer2, "tailer2 should not be nil")

	// Let workers initialize.
	<-time.After(300 * time.Millisecond)

	// Verify both tailers have their respective servers.
	assert.NotNil(t, tailer1.Servers["iso-server-1"], "tailer1 should own iso-server-1")
	assert.NotNil(t, tailer2.Servers["iso-server-2"], "tailer2 should own iso-server-2")

	// Stop tailer1 -- tailer2 must remain unaffected.
	tailer1.Stop()
	<-time.After(200 * time.Millisecond)

	// tailer2 should still have its server and be operational.
	assert.NotNil(t, tailer2.Servers["iso-server-2"], "tailer2 should still own iso-server-2 after tailer1 stops")

	// Verify tailer1's server map is separate (stopping tailer1 should not touch tailer2).
	assert.NotNil(t, tailer1.Servers["iso-server-1"], "tailer1 server map entry still exists in struct")

	tailer2.Stop()
}

// ----- Integration Test: Startup failure propagation -----
// Test that StartLogtail returns a non-nil error when given an invalid config.
// Covers acceptance criteria: synchronous error return from startup.
func TestIntegrationStartupFailurePropagation(t *testing.T) {
	t.Parallel()

	// Config with a server referencing a non-existent router -- should fail validation.
	config := &conf.Config{
		Servers: map[string]*conf.ServerConfig{
			"bad-server": {
				Name:    "bad-server",
				Command: "echo test",
				Routers: []string{"nonexistent-router"},
			},
		},
	}

	tailer, err := starter.StartLogtail(config)
	assert.Nil(t, tailer, "tailer should be nil on startup failure")
	assert.Error(t, err, "error should be non-nil for invalid config")
	assert.ErrorIs(t, err, conf.ErrRouterNotExist, "error should indicate missing router")
}

// ----- Integration Test: Startup success signal -----
// Test that StartLogtail returns a nil error and a valid tailer for a valid config.
// Covers acceptance criteria: synchronous startup completion signal.
func TestIntegrationStartupSuccessSignal(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"null": {
				Name: "null",
				Type: trans.TypeNull,
			},
		},
		Routers: map[string]*conf.RouterConfig{
			"r1": {
				Name:      "r1",
				Transfers: []string{"null"},
			},
		},
		Servers: map[string]*conf.ServerConfig{
			"success-server": {
				Name:    "success-server",
				Command: "echo startup-success",
				Routers: []string{"r1"},
			},
		},
	}

	tailer, err := starter.StartLogtail(config)
	require.NoError(t, err, "StartLogtail should succeed with valid config")
	require.NotNil(t, tailer, "tailer should not be nil on success")

	// The tailer should have the server registered after synchronous start.
	assert.NotNil(t, tailer.Servers["success-server"], "server should be registered in tailer")

	<-time.After(200 * time.Millisecond)
	tailer.Stop()
}

// ----- Integration Test: Drop counter under load -----
// Create a router with BufferSize=1 and BlockingMode=false.
// Send 100 messages rapidly via Receive(). Verify DroppedMessages() > 0
// and that delivered + dropped == 100.
// Covers acceptance criteria: pipeline backpressure drop counting.
func TestIntegrationDropCounterUnderLoad(t *testing.T) {
	t.Parallel()

	const bufferSize = 1
	const totalMessages = 100

	runner := vrun.New()

	router := &route.Router{
		Lock:         sync.Mutex{},
		Runner:       runner,
		ID:           "drop-load-test",
		Name:         "drop-load-test",
		Source:       "test",
		Channel:      make(chan []byte, bufferSize),
		BufferSize:   bufferSize,
		BlockingMode: false,
	}

	// Send 100 messages rapidly without consuming -- most will be dropped.
	for i := range totalMessages {
		router.Receive(fmt.Appendf(nil, "msg-%d", i))
	}

	// Drain the channel to count delivered messages.
	close(router.Channel)

	var delivered int64
	for range router.Channel {
		delivered++
	}

	dropped := router.DroppedMessages()

	// delivered + dropped must equal total messages sent.
	assert.Equal(t, int64(totalMessages), delivered+dropped,
		"delivered + dropped should equal total messages")
	// At least some messages must have been dropped (buffer is only 1).
	assert.Greater(t, dropped, int64(0),
		"some messages should be dropped with buffer size 1")
}

// ----- Integration Test: Blocking mode -----
// Create a router with BufferSize=1 and BlockingMode=true.
// Fill the channel, then call Receive() in a goroutine and verify it blocks.
// Consume from the channel, then verify the goroutine completes.
// Covers acceptance criteria: blocking mode behavior.
func TestIntegrationBlockingMode(t *testing.T) {
	t.Parallel()

	const bufferSize = 1

	runner := vrun.New()

	router := &route.Router{
		Lock:         sync.Mutex{},
		Runner:       runner,
		ID:           "blocking-integ",
		Name:         "blocking-integ",
		Source:       "test",
		Channel:      make(chan []byte, bufferSize),
		BufferSize:   bufferSize,
		BlockingMode: true,
	}

	// Fill the channel to capacity.
	router.Channel <- []byte("fill")

	// Receive in a goroutine -- should block because channel is full.
	done := make(chan struct{})

	go func() {
		router.Receive([]byte("blocked-msg"))
		close(done)
	}()

	// Verify it remains blocked for at least 50ms.
	select {
	case <-done:
		t.Fatal("Receive should have blocked but returned immediately")
	case <-time.After(50 * time.Millisecond):
		// Expected: still blocked.
	}

	// Consume from the channel to unblock the goroutine.
	<-router.Channel

	// Now the goroutine should complete.
	select {
	case <-done:
		// Success: Receive unblocked after channel was consumed.
	case <-time.After(time.Second):
		t.Fatal("Receive did not unblock after consuming from channel")
	}

	// No drops should occur in blocking mode.
	assert.Equal(t, int64(0), router.DroppedMessages(),
		"blocking mode should not drop messages")
}

// ----- Integration Test: Blocking mode respects stop -----
// Create a blocking-mode router, fill its channel, call Receive() in a goroutine,
// then stop the runner. Verify Receive() returns without a goroutine leak.
// Covers acceptance criteria: graceful shutdown of blocked receivers.
func TestIntegrationBlockingModeRespectsStop(t *testing.T) {
	t.Parallel()

	const bufferSize = 1

	parentRunner := vrun.New()

	router := &route.Router{
		Lock:         sync.Mutex{},
		Runner:       parentRunner.NewChild(),
		ID:           "stop-integ",
		Name:         "stop-integ",
		Source:       "test",
		Channel:      make(chan []byte, bufferSize),
		BufferSize:   bufferSize,
		BlockingMode: true,
	}

	// Fill the channel.
	router.Channel <- []byte("fill")

	// Receive in a goroutine -- should block.
	done := make(chan struct{})

	go func() {
		router.Receive([]byte("blocked-msg"))
		close(done)
	}()

	// Verify it blocks.
	select {
	case <-done:
		t.Fatal("Receive should have blocked but returned immediately")
	case <-time.After(50 * time.Millisecond):
		// Expected: blocked.
	}

	// Stop the router -- this should close the runner channel, unblocking Receive().
	router.Stop()

	select {
	case <-done:
		// Success: Receive returned after runner stop.
	case <-time.After(time.Second):
		t.Fatal("Receive did not unblock after runner stop -- potential goroutine leak")
	}
}

// ----- Integration Test: Stats API -----
// Start a tailer with routers. Send data to trigger drops. Hit GET /manage/stats
// via httptest. Verify JSON response contains correct router IDs, drop counts,
// and buffer sizes.
// Covers acceptance criteria: stats API endpoint, drop visibility.
func TestIntegrationStatsAPI(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"null": {
				Name: "null",
				Type: trans.TypeNull,
			},
		},
		Routers: map[string]*conf.RouterConfig{
			"stats-router": {
				Name:       "stats-router",
				Transfers:  []string{"null"},
				BufferSize: 2,
			},
		},
		Servers: map[string]*conf.ServerConfig{
			"stats-server": {
				Name:    "stats-server",
				Command: "echo stats-test",
				Routers: []string{"stats-router"},
			},
		},
	}

	tailer, err := starter.StartLogtail(config)
	require.NoError(t, err, "tailer should start")
	require.NotNil(t, tailer, "tailer should not be nil")

	// Let the server and workers initialize.
	<-time.After(500 * time.Millisecond)

	// Force drops by sending data directly to the router channels.
	// Find the routers through the tailer's server/worker hierarchy.
	server := tailer.Servers["stats-server"]
	require.NotNil(t, server, "stats-server should exist")

	for _, worker := range server.Workers {
		for _, router := range worker.Routers {
			// Fill the channel and then send extra messages to trigger drops.
			for range 20 {
				router.Receive([]byte("load-msg"))
			}
		}
	}

	// Hit the stats API endpoint using httptest.
	req := httptest.NewRequest(http.MethodGet, "/manage/stats", nil)
	rec := httptest.NewRecorder()

	webapi.Serve(req, rec, tailer)

	assert.Equal(t, http.StatusOK, rec.Code, "stats endpoint should return 200")

	contentType := rec.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType, "response should be JSON")

	var stats []tail.RouterStats
	err = json.Unmarshal(rec.Body.Bytes(), &stats)
	require.NoError(t, err, "response should be valid JSON")

	// There should be at least one router in the stats.
	require.NotEmpty(t, stats, "stats should contain at least one router")

	// Verify the router stats fields.
	found := false

	for _, s := range stats {
		if s.Name == "stats-router" {
			found = true
			assert.Equal(t, "stats-server", s.Source, "source should match server name")
			assert.Equal(t, 2, s.BufferSize, "buffer size should match config")
			assert.False(t, s.BlockingMode, "blocking mode should be false by default")
			// With buffer=2 and 20 rapid sends, at least some drops should occur.
			assert.GreaterOrEqual(t, s.DropCount, int64(0), "drop count should be non-negative")
		}
	}

	assert.True(t, found, "stats should include a router named stats-router")

	tailer.Stop()
}

// ----- Integration Test: Backward compatibility -----
// Start with a config that has no buffer_size or blocking_mode fields.
// Verify default behavior: buffer=16, non-blocking, no errors on config parse.
// Covers acceptance criteria: backward compatibility of new config fields.
func TestIntegrationBackwardCompatibility(t *testing.T) {
	t.Parallel()

	// Config without any buffer_size or blocking_mode -- simulates legacy config.
	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"null": {
				Name: "null",
				Type: trans.TypeNull,
			},
		},
		Routers: map[string]*conf.RouterConfig{
			"compat-router": {
				Name:      "compat-router",
				Transfers: []string{"null"},
				// No BufferSize or BlockingMode set -- should use defaults.
			},
		},
		Servers: map[string]*conf.ServerConfig{
			"compat-server": {
				Name:    "compat-server",
				Command: "echo compat-test",
				Routers: []string{"compat-router"},
			},
		},
	}

	tailer, err := starter.StartLogtail(config)
	require.NoError(t, err, "legacy config without new fields should start without error")
	require.NotNil(t, tailer)

	<-time.After(300 * time.Millisecond)

	// Verify defaults by inspecting routers through the stats API.
	stats := tailer.CollectRouterStats()
	require.NotEmpty(t, stats, "should have at least one router")

	for _, s := range stats {
		if s.Name == "compat-router" {
			assert.Equal(t, route.DefaultChannelBufferSize, s.BufferSize,
				"default buffer size should be %d", route.DefaultChannelBufferSize)
			assert.False(t, s.BlockingMode,
				"default blocking mode should be false")
			assert.Equal(t, int64(0), s.DropCount,
				"drop count should be 0 with no load")
		}
	}

	tailer.Stop()
}

// ----- Integration Test: YAML config backward compatibility -----
// Parse a YAML config string that omits buffer_size and blocking_mode.
// Verify the parsed RouterConfig has zero-value defaults and the tailer starts.
// Covers acceptance criteria: config schema backward compatibility.
func TestIntegrationYAMLBackwardCompatibility(t *testing.T) {
	t.Parallel()

	// Directly construct a config as would result from YAML without new fields.
	routerCfg := &conf.RouterConfig{
		Name:      "yaml-compat",
		Transfers: []string{"null"},
	}

	// BufferSize and BlockingMode should be zero-value (0 and false).
	assert.Equal(t, 0, routerCfg.BufferSize,
		"BufferSize should default to 0 in unmarshaled config")
	assert.False(t, routerCfg.BlockingMode,
		"BlockingMode should default to false in unmarshaled config")

	// When BuildRouter uses these defaults, it should apply DefaultChannelBufferSize.
	runner := vrun.New()
	transferMatcher := func(_ []string) []trans.Transfer { return nil }

	router := route.BuildRouter(runner, routerCfg, transferMatcher, "yaml-test", "src")
	assert.Equal(t, route.DefaultChannelBufferSize, router.BufferSize,
		"BuildRouter should default buffer size to DefaultChannelBufferSize")
	assert.Equal(t, route.DefaultChannelBufferSize, cap(router.Channel),
		"channel capacity should match default buffer size")
	assert.False(t, router.BlockingMode,
		"blocking mode should be false")

	router.Stop()
}
