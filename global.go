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
	"sync"

	"github.com/gorilla/websocket"
)

var (
	dingTextMessageDataPrefix = []byte(`{"msgtype":"text","text":{"content":"[logtail-`)
	dingTextMessageDataSuffix = []byte(`"}}`)
	larkTextMessageDataPrefix = []byte(`{"msg_type":"text","content":{"text":"[logtail-`)
	larkTextMessageDataSuffix = []byte(`"}}`)
	messageTitleContentSplit  = []byte("]: ")
	quotationBytes            = []byte(`"`)
	escapeQuotationBytes      = []byte(`\"`)
)

var websocketUpgrader = websocket.Upgrader{}

var (
	serverDBLock = sync.Mutex{}
	serverDB     = make(map[string]*Server, 4)
)

var defaultFormat *Format
