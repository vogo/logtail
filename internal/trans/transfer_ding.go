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
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
)

const TypeDing = "ding"

const (
	dingMessageDataFixedBytesNum = 5
	dingMessageDataMaxLength     = 1024
)

var (
	// nolint:gochecknoglobals // ignore this
	dingTextMessageDataPrefix = []byte(`{"msgtype":"text","text":{"content":"[logtail-`)

	// nolint:gochecknoglobals // ignore this
	dingTextMessageDataSuffix = []byte(`"}}`)

	// nolint:gochecknoglobals // ignore this
	messageTitleContentSplit = []byte("]: ")
)

// transfer next message after the interval, ignore messages in the interval.
const dingMessageTransferInterval = time.Second * 5

type DingTransfer struct {
	Counter
	id           string
	url          string
	prefix       []byte
	transferring int32 // whether transferring message
}

func (d *DingTransfer) Name() string {
	return d.id
}
func (d *DingTransfer) Start() error { return nil }

func (d *DingTransfer) Stop() error { return nil }

// Trans transfer data to dingding.
func (d *DingTransfer) Trans(source string, data ...[]byte) error {
	d.CountIncr()

	if !atomic.CompareAndSwapInt32(&d.transferring, 0, 1) {
		// ignore message to
		return nil
	}

	go func() {
		<-time.After(dingMessageTransferInterval)
		atomic.StoreInt32(&d.transferring, 0)
	}()

	if countMessage, ok := d.CountStat(); ok {
		_ = d.execTrans(source, []byte(countMessage))
	}

	return d.execTrans(source, data...)
}

// nolint:dupl // ignore duplicated code for easy maintenance for diff transfers.
func (d *DingTransfer) execTrans(source string, data ...[]byte) error {
	size := dingMessageDataFixedBytesNum + len(data)
	list := make([][]byte, size)
	list[0] = dingTextMessageDataPrefix
	list[1] = d.prefix
	list[2] = []byte(source)
	list[3] = messageTitleContentSplit

	idx := 4
	messageRemainCapacity := dingMessageDataMaxLength

	for _, bytes := range data {
		if messageRemainCapacity <= 0 {
			break
		}

		bytes = EscapeLimitJSONBytes(bytes, messageRemainCapacity)

		list[idx] = bytes
		idx++

		messageRemainCapacity -= len(bytes)
	}

	list[idx] = dingTextMessageDataSuffix

	if err := httpTrans(d.url, list[:idx+1]...); err != nil {
		logger.Errorf("ding error: %v", err)
	}

	return nil
}

// NewDingTransfer new dingding trans.
func NewDingTransfer(id, url, prefix string) *DingTransfer {
	trans := &DingTransfer{
		id:           id,
		url:          url,
		transferring: 0,
	}

	if prefix == "" {
		prefix = DefaultTransferPrefix
	}

	trans.prefix = []byte(prefix)

	trans.CountReset()

	return trans
}
