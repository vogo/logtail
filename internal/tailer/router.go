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

package tailer

import (
	"fmt"
	"sync"
	"time"

	"github.com/vogo/grunner"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/trans"
)

const DurationReadNextTimeout = time.Millisecond * 60

type Router struct {
	ID        string
	Name      string
	Lock      sync.Mutex
	Runner    *grunner.Runner
	Matchers  []match.Matcher
	Transfers []trans.Transfer
}

func NewRouter(server *Server, name string, matchers []match.Matcher, transfers []trans.Transfer) *Router {
	id := fmt.Sprintf("%s-%s", server.ID, name)

	router := &Router{
		ID:        id,
		Name:      name,
		Lock:      sync.Mutex{},
		Runner:    server.Gorunner.NewChild(),
		Matchers:  matchers,
		Transfers: transfers,
	}

	return router
}

func (r *Router) SetMatchers(matchers []match.Matcher) {
	r.Matchers = matchers
}

func (r *Router) Start() error {
	return nil
}

func (r *Router) Stop() {
	r.Runner.Stop()
}
