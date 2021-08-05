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
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
)

const TransferTypeDing = "ding"

const (
	dingMessageDataFixedBytesNum = 4
	dingMessageDataMaxLength     = 1024
)

// transfer next message after the interval, ignore messages in the interval.
const dingMessageTransferInterval = time.Second * 5

type DingTransfer struct {
	url          string
	transferring int32 // whether transferring message
}

func (d *DingTransfer) start(*Router) error { return nil }

func (d *DingTransfer) Trans(serverID string, data ...[]byte) error {
	if !atomic.CompareAndSwapInt32(&d.transferring, 0, 1) {
		// ignore message to
		return nil
	}

	go func() {
		<-time.After(dingMessageTransferInterval)
		atomic.StoreInt32(&d.transferring, 0)
	}()

	size := dingMessageDataFixedBytesNum + len(data)
	list := make([][]byte, size)
	list[0] = dingTextMessageDataPrefix
	list[1] = []byte(serverID)
	list[2] = messageTitleContentSplit

	idx := 3
	messageRemainCapacity := dingMessageDataMaxLength

	for _, b := range data {
		if messageRemainCapacity <= 0 {
			break
		}

		b = EscapeLimitJSONBytes(b, messageRemainCapacity)

		list[idx] = b
		idx++

		messageRemainCapacity -= len(b)
	}

	list[idx] = dingTextMessageDataSuffix

	if err := httpTrans(d.url, list[:idx+1]...); err != nil {
		logger.Errorf("ding error: %v", err)
	}

	return nil
}

func NewDingTransfer(url string) Transfer {
	return &DingTransfer{
		url:          url,
		transferring: 0,
	}
}
