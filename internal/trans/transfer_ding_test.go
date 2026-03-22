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
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/trans"
)

func TestDingTransferWithRateLimit(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := trans.NewDingTransfer("ding-rl", server.URL, "test-", trans.HTTPTransferOptions{
		RateLimit: 1.0,
		RateBurst: 1,
	})
	defer func() { _ = dt.Stop() }()

	// First call goes through the CAS + rate limiter
	err := dt.Trans("src", []byte("msg1"))
	require.NoError(t, err)

	// The first Trans triggers execTrans directly and sets transferring=1.
	// Subsequent calls are blocked by the CAS, not the rate limiter.
	// So we only expect 1 request from the initial call.
	assert.GreaterOrEqual(t, int(requestCount.Load()), 1)
}

func TestDingTransferNoRateLimit(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := trans.NewDingTransfer("ding-norl", server.URL, "test-", trans.HTTPTransferOptions{})
	defer func() { _ = dt.Stop() }()

	err := dt.Trans("src", []byte("msg1"))
	require.NoError(t, err)

	// Without rate limiting, the first message should go through
	assert.Equal(t, int32(1), requestCount.Load())
}

func TestDingTransferName(t *testing.T) {
	t.Parallel()

	dt := trans.NewDingTransfer("my-ding", "http://example.com", "", trans.HTTPTransferOptions{})
	defer func() { _ = dt.Stop() }()

	assert.Equal(t, "my-ding", dt.Name())
}
