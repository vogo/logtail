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
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/trans"
)

func TestWebhookTransferNoBatching(t *testing.T) {
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

	wh := trans.NewWebhookTransfer("wh-nobatch", server.URL, "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	require.NoError(t, wh.Trans("src", []byte("msg1")))
	require.NoError(t, wh.Trans("src", []byte("msg2")))

	mu.Lock()
	defer mu.Unlock()

	assert.Len(t, requests, 2)
	assert.Equal(t, "msg1", requests[0])
	assert.Equal(t, "msg2", requests[1])
}

func TestWebhookTransferWithBatching(t *testing.T) {
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

	wh := trans.NewWebhookTransfer("wh-batch", server.URL, "", trans.HTTPTransferOptions{
		BatchSize:    3,
		BatchTimeout: 10 * time.Second,
	})
	defer func() { _ = wh.Stop() }()

	require.NoError(t, wh.Trans("src", []byte("msg1")))
	require.NoError(t, wh.Trans("src", []byte("msg2")))
	require.NoError(t, wh.Trans("src", []byte("msg3")))

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, requests, 1)
	assert.Equal(t, "msg1\nmsg2\nmsg3", requests[0])
}

func TestWebhookTransferBatchTimeout(t *testing.T) {
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

	wh := trans.NewWebhookTransfer("wh-timeout", server.URL, "", trans.HTTPTransferOptions{
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
	})
	defer func() { _ = wh.Stop() }()

	require.NoError(t, wh.Trans("src", []byte("msg1")))
	require.NoError(t, wh.Trans("src", []byte("msg2")))

	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, requests, 1)
	assert.Equal(t, "msg1\nmsg2", requests[0])
}

func TestWebhookTransferBatchFlushOnStop(t *testing.T) {
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

	wh := trans.NewWebhookTransfer("wh-stop", server.URL, "", trans.HTTPTransferOptions{
		BatchSize:    100,
		BatchTimeout: 10 * time.Second,
	})

	require.NoError(t, wh.Trans("src", []byte("msg1")))
	require.NoError(t, wh.Trans("src", []byte("msg2")))

	_ = wh.Stop()

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, requests, 1)
	assert.Equal(t, "msg1\nmsg2", requests[0])
}

func TestWebhookTransferName(t *testing.T) {
	t.Parallel()

	wh := trans.NewWebhookTransfer("my-webhook", "http://example.com", "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	assert.Equal(t, "my-webhook", wh.Name())
}

func TestWebhookTransferBackwardCompatibility(t *testing.T) {
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

	// Zero-value opts: no batching, no rate limiting
	wh := trans.NewWebhookTransfer("wh-compat", server.URL, "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	for i := 0; i < 5; i++ {
		require.NoError(t, wh.Trans("src", []byte("msg")))
	}

	mu.Lock()
	defer mu.Unlock()

	// Each Trans() should produce one HTTP request
	assert.Len(t, requests, 5)
}
