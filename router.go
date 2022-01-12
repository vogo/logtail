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
	"fmt"
	"sync"
	"time"

	"github.com/vogo/grunner"
	"github.com/vogo/logtail/transfer"
)

const DurationReadNextTimeout = time.Millisecond * 60

type Router struct {
	id        string
	Name      string
	lock      sync.Mutex
	Runner    *grunner.Runner
	matchers  []Matcher
	transfers []transfer.Transfer
}

func NewRouter(server *Server, name string, matchers []Matcher, transfers []transfer.Transfer) *Router {
	id := fmt.Sprintf("%s-%s", server.id, name)

	router := &Router{
		id:        id,
		Name:      name,
		lock:      sync.Mutex{},
		Runner:    server.gorunner.NewChild(),
		matchers:  matchers,
		transfers: transfers,
	}

	return router
}

func (r *Router) SetMatchers(matchers []Matcher) {
	r.matchers = matchers
}

func (r *Router) Start() error {
	return nil
}

func (r *Router) Stop() {
	r.Runner.Stop()
}
