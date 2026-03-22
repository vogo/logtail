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
	"errors"
	"net/http"
)

const TypeWebhook = "webhook"

var ErrHTTPStatusNonOK = errors.New("http status non ok")

type WebhookTransfer struct {
	id      string
	url     string
	prefix  string
	client  *http.Client
	batcher *Batcher // nil when batch_size <= 1
}

func (d *WebhookTransfer) Name() string {
	return d.id
}

func (d *WebhookTransfer) Trans(source string, data ...[]byte) error {
	if d.batcher != nil {
		for _, b := range data {
			d.batcher.Add(source, b)
		}

		return nil
	}

	return httpTransWithClient(d.client, d.url, data...)
}

func (d *WebhookTransfer) Start() error { return nil }

func (d *WebhookTransfer) Stop() error {
	if d.batcher != nil {
		d.batcher.Stop()
	}

	closeHTTPClient(d.client)

	return nil
}

// NewWebhookTransfer new webhook trans.
func NewWebhookTransfer(id, url, prefix string, opts HTTPTransferOptions) *WebhookTransfer {
	t := &WebhookTransfer{
		id:     id,
		url:    url,
		prefix: prefix,
		client: NewHTTPClient(HTTPClientConfig{
			MaxIdleConnsPerHost: opts.MaxIdleConnsPerHost,
			IdleConnTimeout:     opts.IdleConnTimeout,
		}),
	}

	if t.prefix == "" {
		t.prefix = DefaultTransferPrefix
	}

	if opts.BatchSize > 1 {
		t.batcher = NewBatcher(opts.BatchSize, opts.BatchTimeout, func(source string, data []byte) error {
			return httpTransWithClient(t.client, t.url, data)
		})
	}

	return t
}
