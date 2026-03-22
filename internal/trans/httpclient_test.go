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
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/trans"
)

func TestNewHTTPClientDefaults(t *testing.T) {
	t.Parallel()

	client := trans.NewHTTPClient(trans.HTTPClientConfig{})

	assert.NotNil(t, client)
	assert.NotNil(t, client.Transport)

	tp, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 2, tp.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, tp.IdleConnTimeout)
}

func TestNewHTTPClientCustomValues(t *testing.T) {
	t.Parallel()

	client := trans.NewHTTPClient(trans.HTTPClientConfig{
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	})

	tp, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 10, tp.MaxIdleConnsPerHost)
	assert.Equal(t, 30*time.Second, tp.IdleConnTimeout)
}

func TestHTTPClientConnectionReuse(t *testing.T) {
	t.Parallel()

	var connCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connCount.Add(1)
		}
	}

	wh := trans.NewWebhookTransfer("test", server.URL, "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	for i := 0; i < 10; i++ {
		err := wh.Trans("src", []byte("hello"))
		require.NoError(t, err)
	}

	// Connections should be reused, so we expect very few new connections.
	assert.LessOrEqual(t, int(connCount.Load()), 2)
}

func TestWebhookTransferPostSuccess(t *testing.T) {
	t.Parallel()

	var received []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := trans.NewWebhookTransfer("test-wh", server.URL, "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	err := wh.Trans("source1", []byte("test-data"))
	require.NoError(t, err)
	assert.Equal(t, "test-data", string(received))
}

func TestWebhookTransferPostError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	wh := trans.NewWebhookTransfer("test-err", server.URL, "", trans.HTTPTransferOptions{})
	defer func() { _ = wh.Stop() }()

	err := wh.Trans("source1", []byte("test-data"))
	assert.Error(t, err)
}

func TestClientIsolation(t *testing.T) {
	t.Parallel()

	var count1, count2 atomic.Int32

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count1.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count2.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	wh1 := trans.NewWebhookTransfer("wh1", server1.URL, "", trans.HTTPTransferOptions{})
	wh2 := trans.NewWebhookTransfer("wh2", server2.URL, "", trans.HTTPTransferOptions{})

	// Send to both
	require.NoError(t, wh1.Trans("s", []byte("a")))
	require.NoError(t, wh2.Trans("s", []byte("b")))

	// Stop wh1
	_ = wh1.Stop()

	// wh2 should still work
	require.NoError(t, wh2.Trans("s", []byte("c")))

	_ = wh2.Stop()

	assert.Equal(t, int32(1), count1.Load())
	assert.Equal(t, int32(2), count2.Load())
}
