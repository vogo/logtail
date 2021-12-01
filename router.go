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

	"github.com/vogo/gstop"
)

const DurationReadNextTimeout = time.Millisecond * 60

type Router struct {
	id        int64
	name      string
	lock      sync.Mutex
	stopper   *gstop.Stopper
	matchers  []Matcher
	transfers []Transfer
}

func NewRouter(s *Server, matchers []Matcher, transfers []Transfer) *Router {
	routerID := s.nextRouterID()
	name := fmt.Sprintf("%s-%d", s.id, routerID)

	t := &Router{
		id:        routerID,
		name:      name,
		lock:      sync.Mutex{},
		stopper:   s.stopper.NewChild(),
		matchers:  matchers,
		transfers: transfers,
	}

	return t
}

func (r *Router) SetMatchers(matchers []Matcher) {
	r.matchers = matchers
}

func (r *Router) Start() error {
	return nil
}

func (r *Router) Stop() {
	r.stopper.Stop()
}
