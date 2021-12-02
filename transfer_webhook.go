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

package logtail

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vogo/logger"
	"github.com/vogo/vogo/vio"
)

const TransferTypeWebhook = "webhook"

var ErrHTTPStatusNonOK = errors.New("http status non ok")

type WebhookTransfer struct {
	id  string
	url string
}

func (d *WebhookTransfer) ID() string {
	return d.id
}

func (d *WebhookTransfer) Trans(_ string, data ...[]byte) error {
	return httpTrans(d.url, data...)
}

func (d *WebhookTransfer) Start() error { return nil }

func (d *WebhookTransfer) Stop() error { return nil }

// NewWebhookTransfer new webhook transfer.
func NewWebhookTransfer(id, url string) *WebhookTransfer {
	return &WebhookTransfer{
		id:  id,
		url: url,
	}
}

func httpTrans(url string, data ...[]byte) error {
	res, err := http.Post(url, "application/json", vio.NewBytesReader(data...))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			logger.Warnf("http alert error: %s", body)
		}

		return fmt.Errorf("http alert error, %w: %d", ErrHTTPStatusNonOK, res.StatusCode)
	}

	return nil
}
