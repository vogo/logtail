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

package route

import (
	"sync"

	"github.com/vogo/gorun"
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/trans"
)

const DefaultChannelBufferSize = 16

type RoutersBuilder func() *[]Router

type Router struct {
	Lock      sync.Mutex
	Runner    *gorun.Runner
	ID        string
	Name      string
	Source    string
	Channel   chan []byte
	Matchers  []match.Matcher
	Transfers []trans.Transfer
}

func BuildRouter(workerRunner *gorun.Runner,
	routerConfig *conf.RouterConfig,
	transfersFunc trans.TransferMatcher,
	routerID string, source string,
) *Router {
	matchers, err := NewMatchers(routerConfig.Matchers)
	if err != nil {
		panic(err)
	}

	router := &Router{
		ID:        routerID,
		Name:      routerConfig.Name,
		Source:    source,
		Lock:      sync.Mutex{},
		Runner:    workerRunner.NewChild(),
		Channel:   make(chan []byte, DefaultChannelBufferSize),
		Matchers:  matchers,
		Transfers: transfersFunc(routerConfig.Transfers),
	}

	return router
}

func (r *Router) SetMatchers(matchers []match.Matcher) {
	r.Matchers = matchers
}

// Route match lines and transfer.
func (r *Router) Route(data []byte) error {
	if len(r.Matchers) == 0 {
		return r.Trans(data)
	}

	if !r.Matches(data) {
		return nil
	}

	return r.Trans(data)
}

func (r *Router) Trans(data []byte) error {
	transfers := r.Transfers
	if len(transfers) == 0 {
		return nil
	}

	for _, t := range transfers {
		if err := t.Trans(r.Source, data); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) Stop() {
	r.Runner.StopWith(func() {
		logger.Infof("Routers [%s] stopping", r.ID)
		close(r.Channel)
	})
}

func (r *Router) Receive(data []byte) {
	defer func() {
		_ = recover()
	}()

	select {
	case <-r.Runner.C:
		return
	case r.Channel <- data:
	default:
	}
}

func (r *Router) Matches(bytes []byte) bool {
	for _, m := range r.Matchers {
		if !m.Match(bytes) {
			return false
		}
	}

	return true
}
