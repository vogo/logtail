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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/route"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/vogo/vsync/vrun"
)

func TestRoute(t *testing.T) {
	t.Parallel()

	router := &route.Router{
		Lock:    sync.Mutex{},
		Runner:  vrun.New(),
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

	//nolint:lll //ignore this.
	testLogMessage := `2022-05-20 13:35:53.794 ERROR [ConsumeMessageThread_1] [-] h.t.c.c.i.Service - 发起失败!失败原因msg=订单不存在, 参数postMap={"data":"xxx","open_id":"d99bcfde2e727e5eee7d4b5488741234","open_key":"17b130ef168500054f02b814bf261234","sign":"03f17f6f57235a8e0181aaa71ef51234","timestamp":"1653024953"}, 返回结果result={"errcode":8018,"msg":"\u539f\u59cb\u8ba2\u5355\u4e0d\u5b58\u5728"}`

	err := router.Route([]byte(testLogMessage))

	assert.Nil(t, err)
}

func TestReceiveDropCounting(t *testing.T) {
	t.Parallel()

	const bufferSize = 1
	const totalMessages = 100

	runner := vrun.New()

	router := &route.Router{
		Lock:         sync.Mutex{},
		Runner:       runner,
		ID:           "drop-test",
		Name:         "drop-test",
		Source:       "test",
		Channel:      make(chan []byte, bufferSize),
		BufferSize:   bufferSize,
		BlockingMode: false,
	}

	// Send messages rapidly without consuming
	for i := 0; i < totalMessages; i++ {
		router.Receive([]byte("msg"))
	}

	// Drain the channel to count delivered messages
	close(router.Channel)

	var delivered int64
	for range router.Channel {
		delivered++
	}

	dropped := router.DroppedMessages()
	assert.Equal(t, int64(totalMessages), delivered+dropped)
	assert.Greater(t, dropped, int64(0))
}

func TestReceiveBlockingMode(t *testing.T) {
	t.Parallel()

	const bufferSize = 1

	runner := vrun.New()

	router := &route.Router{
		Lock:         sync.Mutex{},
		Runner:       runner,
		ID:           "blocking-test",
		Name:         "blocking-test",
		Source:       "test",
		Channel:      make(chan []byte, bufferSize),
		BufferSize:   bufferSize,
		BlockingMode: true,
	}

	// Fill the channel
	router.Channel <- []byte("fill")

	// Receive in a goroutine -- should block since channel is full
	done := make(chan struct{})

	go func() {
		router.Receive([]byte("blocked"))
		close(done)
	}()

	// Verify it blocks for a short period
	select {
	case <-done:
		t.Fatal("Receive should have blocked but returned immediately")
	case <-time.After(50 * time.Millisecond):
		// expected: still blocked
	}

	// Consume from the channel to unblock
	<-router.Channel

	// Now the goroutine should complete
	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("Receive did not unblock after consuming from channel")
	}

	// No drops in blocking mode
	assert.Equal(t, int64(0), router.DroppedMessages())
}

func TestBlockingModeRespectsStop(t *testing.T) {
	t.Parallel()

	const bufferSize = 1

	runner := vrun.New()

	router := &route.Router{
		Lock:         sync.Mutex{},
		Runner:       runner.NewChild(),
		ID:           "stop-test",
		Name:         "stop-test",
		Source:       "test",
		Channel:      make(chan []byte, bufferSize),
		BufferSize:   bufferSize,
		BlockingMode: true,
	}

	// Fill the channel
	router.Channel <- []byte("fill")

	// Receive in a goroutine -- should block since channel is full
	done := make(chan struct{})

	go func() {
		router.Receive([]byte("blocked"))
		close(done)
	}()

	// Verify it blocks
	select {
	case <-done:
		t.Fatal("Receive should have blocked but returned immediately")
	case <-time.After(50 * time.Millisecond):
	}

	// Stop the runner -- should unblock Receive
	router.Stop()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("Receive did not unblock after runner stop")
	}
}

func TestConfigurableBufferSize(t *testing.T) {
	t.Parallel()

	runner := vrun.New()
	transferMatcher := func(_ []string) []trans.Transfer { return nil }

	// Test custom buffer size
	routerConfig := &conf.RouterConfig{
		Name:       "custom-buf",
		BufferSize: 64,
	}

	router := route.BuildRouter(runner, routerConfig, transferMatcher, "buf-test", "source")
	assert.Equal(t, 64, router.BufferSize)
	assert.Equal(t, 64, cap(router.Channel))

	router.Stop()

	// Test default buffer size when zero
	routerConfig2 := &conf.RouterConfig{
		Name:       "default-buf",
		BufferSize: 0,
	}

	router2 := route.BuildRouter(runner, routerConfig2, transferMatcher, "buf-test-2", "source")
	assert.Equal(t, route.DefaultChannelBufferSize, router2.BufferSize)
	assert.Equal(t, route.DefaultChannelBufferSize, cap(router2.Channel))

	router2.Stop()

	// Test default buffer size when negative
	routerConfig3 := &conf.RouterConfig{
		Name:       "neg-buf",
		BufferSize: -1,
	}

	router3 := route.BuildRouter(runner, routerConfig3, transferMatcher, "buf-test-3", "source")
	assert.Equal(t, route.DefaultChannelBufferSize, router3.BufferSize)

	router3.Stop()
}

func TestBuildRouterBlockingMode(t *testing.T) {
	t.Parallel()

	runner := vrun.New()
	transferMatcher := func(_ []string) []trans.Transfer { return nil }

	routerConfig := &conf.RouterConfig{
		Name:         "blocking",
		BlockingMode: true,
		BufferSize:   8,
	}

	router := route.BuildRouter(runner, routerConfig, transferMatcher, "block-test", "source")
	assert.True(t, router.BlockingMode)
	assert.Equal(t, 8, router.BufferSize)

	router.Stop()
}
