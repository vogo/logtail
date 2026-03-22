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

package trans_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/trans"
)

// Test 1: HTTP Connection Reuse
// Verifies that the HTTP client reuses connections across multiple sequential requests,
// rather than creating a new connection per request. A test HTTP server with a ConnState
// hook counts new connections. After 10 sequential requests through a WebhookTransfer,
// the number of new connections should be <= 2.
func TestIntegrationHTTPConnectionReuse(t *testing.T) {
	t.Parallel()

	var connCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Track new connections via ConnState hook.
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connCount.Add(1)
		}
	}

	wh := trans.NewWebhookTransfer("int-conn-reuse", server.URL, "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	// Send 10 messages sequentially.
	for i := range 10 {
		err := wh.Trans("src", fmt.Appendf(nil, "message-%d", i))
		require.NoError(t, err, "Trans call %d should not error", i)
	}

	// Connections should be reused; expect at most 2 new connections.
	assert.LessOrEqual(t, int(connCount.Load()), 2,
		"expected <= 2 new connections due to connection reuse, got %d", connCount.Load())
}

// Test 2: Rate Limiter Drops Messages (DingTransfer)
// Verifies that the rate limiter in DingTransfer drops excess messages. With RateLimit=1.0
// and RateBurst=1, calling Trans() 5 times rapidly should result in only 1-2 HTTP requests
// reaching the server, because the CAS throttle and rate limiter together filter most calls.
func TestIntegrationDingRateLimiterDropsMessages(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := trans.NewDingTransfer("int-ding-rl", server.URL, "test-", trans.HTTPTransferOptions{
		RateLimit: 1.0,
		RateBurst: 1,
	})
	defer func() { _ = dt.Stop() }()

	// Call Trans() 5 times in rapid succession.
	for i := range 5 {
		_ = dt.Trans("src", fmt.Appendf(nil, "msg-%d", i))
	}

	// Allow goroutines to complete.
	time.Sleep(100 * time.Millisecond)

	// Only the first call should get past the CAS guard. The rate limiter allows
	// the first request (1 token available). Subsequent calls are blocked by CAS.
	// So we expect 1-2 actual HTTP requests.
	count := int(requestCount.Load())
	assert.LessOrEqual(t, count, 2,
		"expected at most 2 HTTP requests with rate limiting, got %d", count)
	assert.GreaterOrEqual(t, count, 1,
		"expected at least 1 HTTP request, got %d", count)
}

// Test 3: Batcher Threshold Flush
// Verifies that the batcher flushes when the batch size threshold is reached. With
// BatchSize=3 and BatchTimeout=10s (effectively infinite for this test), calling Trans()
// 3 times should result in exactly 1 HTTP request containing all 3 lines newline-delimited.
func TestIntegrationBatcherThresholdFlush(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var requests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		mu.Lock()
		requests = append(requests, string(body))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := trans.NewWebhookTransfer("int-batch-thresh", server.URL, "", trans.HTTPTransferOptions{
		BatchSize:    3,
		BatchTimeout: 10 * time.Second,
	})
	defer func() { _ = wh.Stop() }()

	// Send exactly 3 messages to trigger threshold flush.
	require.NoError(t, wh.Trans("src", []byte("line-1")))
	require.NoError(t, wh.Trans("src", []byte("line-2")))
	require.NoError(t, wh.Trans("src", []byte("line-3")))

	mu.Lock()
	defer mu.Unlock()

	// Exactly 1 HTTP request should have been made with all 3 lines.
	require.Len(t, requests, 1, "expected exactly 1 HTTP request from threshold flush")
	assert.Equal(t, "line-1\nline-2\nline-3", requests[0],
		"expected all 3 lines newline-delimited in a single request")
}

// Test 4: Batcher Timeout Flush
// Verifies that the batcher flushes when the timeout expires, even if the batch size
// threshold has not been reached. With BatchSize=100 and BatchTimeout=100ms, calling
// Trans() 2 times and waiting 200ms should result in exactly 1 HTTP request with 2 lines.
func TestIntegrationBatcherTimeoutFlush(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var requests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		mu.Lock()
		requests = append(requests, string(body))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := trans.NewWebhookTransfer("int-batch-timeout", server.URL, "", trans.HTTPTransferOptions{
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
	})
	defer func() { _ = wh.Stop() }()

	// Send 2 messages (below threshold of 100).
	require.NoError(t, wh.Trans("src", []byte("line-1")))
	require.NoError(t, wh.Trans("src", []byte("line-2")))

	// Wait for timeout to trigger flush.
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Exactly 1 HTTP request containing the 2 lines.
	require.Len(t, requests, 1, "expected exactly 1 HTTP request from timeout flush")
	assert.Equal(t, "line-1\nline-2", requests[0],
		"expected 2 lines newline-delimited in a single request")
}

// Test 5: Batcher Flush on Stop
// Verifies that calling Stop() on a WebhookTransfer with a batcher flushes any remaining
// buffered entries. With BatchSize=100 and BatchTimeout=10s, calling Trans() 2 times and
// then Stop() should result in exactly 1 HTTP request with the 2 lines.
func TestIntegrationBatcherFlushOnStop(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var requests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		mu.Lock()
		requests = append(requests, string(body))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := trans.NewWebhookTransfer("int-batch-stop", server.URL, "", trans.HTTPTransferOptions{
		BatchSize:    100,
		BatchTimeout: 10 * time.Second,
	})

	// Send 2 messages (below threshold, long timeout).
	require.NoError(t, wh.Trans("src", []byte("line-1")))
	require.NoError(t, wh.Trans("src", []byte("line-2")))

	// Stop flushes remaining data.
	_ = wh.Stop()

	mu.Lock()
	defer mu.Unlock()

	// Exactly 1 HTTP request containing the 2 lines flushed on stop.
	require.Len(t, requests, 1, "expected exactly 1 HTTP request from stop flush")
	assert.Equal(t, "line-1\nline-2", requests[0],
		"expected 2 lines newline-delimited flushed on stop")
}

// Test 6: Backward Compatibility (No Config)
// Verifies that transfers created with zero-value HTTPTransferOptions behave identically
// to the pre-optimization behavior: each Trans() call produces one HTTP request, with
// no rate limiting and no batching.
func TestIntegrationBackwardCompatibility(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var requests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		mu.Lock()
		requests = append(requests, string(body))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Zero-value opts: no batching, no rate limiting.
	wh := trans.NewWebhookTransfer("int-compat", server.URL, "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	// Send 5 messages.
	for i := range 5 {
		err := wh.Trans("src", fmt.Appendf(nil, "msg-%d", i))
		require.NoError(t, err)
	}

	mu.Lock()
	defer mu.Unlock()

	// Each Trans() should produce exactly one HTTP request.
	assert.Len(t, requests, 5,
		"expected 5 individual HTTP requests with zero-value config (backward compatibility)")

	for i := range 5 {
		assert.Equal(t, fmt.Sprintf("msg-%d", i), requests[i])
	}
}

// Test 7: Client Isolation
// Verifies that two WebhookTransfer instances use independent HTTP clients. Stopping one
// transfer should not affect the other's ability to send requests.
func TestIntegrationClientIsolation(t *testing.T) {
	t.Parallel()

	var count1, count2 atomic.Int32

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		count1.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		count2.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	wh1 := trans.NewWebhookTransfer("int-iso-1", server1.URL, "", trans.HTTPTransferOptions{})
	wh2 := trans.NewWebhookTransfer("int-iso-2", server2.URL, "", trans.HTTPTransferOptions{})

	// Send to both.
	require.NoError(t, wh1.Trans("s", []byte("a")))
	require.NoError(t, wh2.Trans("s", []byte("b")))

	// Stop wh1.
	_ = wh1.Stop()

	// wh2 should still work after wh1 is stopped.
	require.NoError(t, wh2.Trans("s", []byte("c")))
	require.NoError(t, wh2.Trans("s", []byte("d")))

	_ = wh2.Stop()

	assert.Equal(t, int32(1), count1.Load(),
		"server1 should have received exactly 1 request before wh1 was stopped")
	assert.Equal(t, int32(3), count2.Load(),
		"server2 should have received 3 requests (wh2 unaffected by wh1 stop)")
}

// Test 8: LarkTransfer Rate Limiting
// Verifies that the rate limiter integrates correctly with LarkTransfer's existing
// 5-second CAS throttle. With RateLimit=5.0 and RateBurst=5, calling Trans() multiple
// times rapidly should be filtered by both the CAS throttle and the rate limiter.
func TestIntegrationLarkRateLimiting(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lt := trans.NewLarkTransfer("int-lark-rl", server.URL, "test-", trans.HTTPTransferOptions{
		RateLimit: 5.0,
		RateBurst: 5,
	})
	defer func() { _ = lt.Stop() }()

	// Call Trans() 10 times rapidly.
	for i := range 10 {
		_ = lt.Trans("src", fmt.Appendf(nil, "lark-msg-%d", i))
	}

	// Allow goroutines to complete.
	time.Sleep(100 * time.Millisecond)

	count := int(requestCount.Load())

	// The CAS throttle means only the first call goes through directly.
	// Subsequent calls are blocked by CAS (transferring=1). The rate limiter
	// with burst=5 has enough tokens for the first call. So we expect a small
	// number of requests (1-2 at most from the initial call + potential count stat).
	assert.GreaterOrEqual(t, count, 1,
		"expected at least 1 HTTP request from LarkTransfer")
	assert.LessOrEqual(t, count, 5,
		"expected at most 5 HTTP requests with rate limiting + CAS throttle")
}

// Test 9: No Duplicate Timer in Batcher
// Verifies that the batcher's nil-timer guard prevents duplicate timers. After adding
// one item and waiting for the timer to fire, then adding another item, exactly 2 flushes
// should occur (one per timer period), not more.
func TestIntegrationBatcherNoDuplicateTimer(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var requestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)

		mu.Lock()
		requestCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := trans.NewWebhookTransfer("int-no-dup-timer", server.URL, "", trans.HTTPTransferOptions{
		BatchSize:    5,
		BatchTimeout: 50 * time.Millisecond,
	})
	defer func() { _ = wh.Stop() }()

	// Add one item, wait for timer flush.
	require.NoError(t, wh.Trans("src", []byte("first")))
	time.Sleep(80 * time.Millisecond)

	// Add another item, wait for timer flush.
	require.NoError(t, wh.Trans("src", []byte("second")))
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Exactly 2 HTTP requests: one per timer flush, no duplicates.
	assert.Equal(t, 2, requestCount,
		"expected exactly 2 HTTP requests (one per timer flush), got %d", requestCount)
}
