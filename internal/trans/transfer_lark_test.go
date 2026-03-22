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

func TestLarkTransferWithRateLimit(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lt := trans.NewLarkTransfer("lark-rl", server.URL, "test-", trans.HTTPTransferOptions{
		RateLimit: 1.0,
		RateBurst: 1,
	})
	defer func() { _ = lt.Stop() }()

	err := lt.Trans("src", []byte("msg1"))
	require.NoError(t, err)

	assert.GreaterOrEqual(t, int(requestCount.Load()), 1)
}

func TestLarkTransferNoRateLimit(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lt := trans.NewLarkTransfer("lark-norl", server.URL, "test-", trans.HTTPTransferOptions{})
	defer func() { _ = lt.Stop() }()

	err := lt.Trans("src", []byte("msg1"))
	require.NoError(t, err)

	assert.Equal(t, int32(1), requestCount.Load())
}

func TestLarkTransferName(t *testing.T) {
	t.Parallel()

	lt := trans.NewLarkTransfer("my-lark", "http://example.com", "", trans.HTTPTransferOptions{})
	defer func() { _ = lt.Stop() }()

	assert.Equal(t, "my-lark", lt.Name())
}
