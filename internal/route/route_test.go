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

package route_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/gorun"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/route"
	"github.com/vogo/logtail/internal/trans"
)

func TestRoute(t *testing.T) {
	t.Parallel()

	router := &route.Router{
		Lock:    sync.Mutex{},
		Runner:  gorun.New(),
		ID:      "test-router",
		Name:    "test-router",
		Source:  "",
		Channel: make(chan []byte, route.DefaultChannelBufferSize),
		Matchers: []match.Matcher{
			match.NewContainsMatcher("ERROR", true),
			match.NewContainsMatcher("参数错误", false),
			match.NewContainsMatcher("不存在", false),
		},
		Transfers: []trans.Transfer{&trans.ConsoleTransfer{}},
	}

	// nolint:lll //ignore this.
	testLogMessage := `2022-05-20 13:35:53.794 ERROR [ConsumeMessageThread_1] [-] h.t.c.c.i.Service - 发起失败!失败原因msg=订单不存在, 参数postMap={"data":"xxx","open_id":"d99bcfde2e727e5eee7d4b5488741234","open_key":"17b130ef168500054f02b814bf261234","sign":"03f17f6f57235a8e0181aaa71ef51234","timestamp":"1653024953"}, 返回结果result={"errcode":8018,"msg":"\u539f\u59cb\u8ba2\u5355\u4e0d\u5b58\u5728"}`

	err := router.Route([]byte(testLogMessage))

	assert.Nil(t, err)
}
