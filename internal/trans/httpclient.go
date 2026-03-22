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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vogo/vogo/vio"
	"github.com/vogo/vogo/vlog"
)

const (
	defaultMaxIdleConnsPerHost = 2
	defaultIdleConnTimeout     = 90 * time.Second
)

// HTTPClientConfig holds HTTP transport configuration.
type HTTPClientConfig struct {
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

// HTTPTransferOptions holds parsed configuration for HTTP-based transfers.
// BuildTransfer in internal/tail parses conf.TransferConfig into this struct.
type HTTPTransferOptions struct {
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	RateLimit           float64       // requests per second; 0 = disabled
	RateBurst           int           // burst size; defaults to 1
	BatchSize           int           // lines per batch; 0 or 1 = disabled
	BatchTimeout        time.Duration // max wait before flush; defaults to 1s
}

// NewHTTPClient creates an *http.Client with a configured transport.
func NewHTTPClient(cfg HTTPClientConfig) *http.Client {
	maxIdle := cfg.MaxIdleConnsPerHost
	if maxIdle <= 0 {
		maxIdle = defaultMaxIdleConnsPerHost
	}

	idleTimeout := cfg.IdleConnTimeout
	if idleTimeout <= 0 {
		idleTimeout = defaultIdleConnTimeout
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost: maxIdle,
		IdleConnTimeout:     idleTimeout,
	}

	return &http.Client{
		Transport: transport,
	}
}

// httpTransWithClient performs an HTTP POST using the provided client.
// It fully drains and closes the response body to ensure connection reuse.
func httpTransWithClient(client *http.Client, url string, data ...[]byte) error {
	res, err := client.Post(url, "application/json", vio.NewBytesReader(data...))
	if err != nil {
		return err
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		if respBody, respErr := io.ReadAll(res.Body); respErr == nil {
			vlog.Warnf("http alert error! response: %s, request data length: %d", respBody, len(data))
		}

		return fmt.Errorf("http alert error, %w: %d", ErrHTTPStatusNonOK, res.StatusCode)
	}

	// Drain response body to enable connection reuse.
	_, _ = io.Copy(io.Discard, res.Body)

	return nil
}

// closeHTTPClient closes idle connections on the client's transport.
func closeHTTPClient(client *http.Client) {
	if client == nil {
		return
	}

	if t, ok := client.Transport.(*http.Transport); ok {
		t.CloseIdleConnections()
	}
}
