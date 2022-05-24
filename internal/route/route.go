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
	"time"

	"github.com/vogo/gorun"
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/logtail/internal/util"
)

const DurationReadNextTimeout = time.Millisecond * 60

type RoutersBuilder func() *[]Router

type Router struct {
	Lock      sync.Mutex
	Runner    *gorun.Runner
	ID        string
	Name      string
	Source    string
	Format    *match.Format
	Channel   Channel
	Matchers  []match.Matcher
	Transfers []trans.Transfer
}

func StartRouter(workerRunner *gorun.Runner,
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
		Channel:   make(Channel),
		Matchers:  matchers,
		Transfers: transfersFunc(routerConfig.Transfers),
	}

	go router.StartLoop()

	return router
}

func (r *Router) SetMatchers(matchers []match.Matcher) {
	r.Matchers = matchers
}

func (r *Router) Route(bytes []byte) error {
	if len(r.Matchers) == 0 {
		return r.Trans(bytes)
	}

	bytes = match.IndexToLineStart(r.Format, bytes)

	var (
		list    [][]byte
		matches []byte
	)

	idx := 0
	length := len(bytes)

	for idx < length {
		matches, idx = r.Match(bytes, length, idx)

		if len(matches) > 0 {
			list = append(list, matches)

			for length > 0 && idx >= length {
				length, idx = r.ReadMoreFollowingLines(&list, &bytes)
			}

			if err := r.Trans(list...); err != nil {
				return err
			}

			list = nil
		}
	}

	return nil
}

// nolint:gocritic //ignore this.
func (r *Router) ReadMoreFollowingLines(list *[][]byte, bytes *[]byte) (int, int) {
	*bytes = r.NextBytes()
	idx := 0
	length := len(*bytes)

	if length > 0 {
		var end int
		// append following lines
		idx, end = IndexFollowingLines(r.Format, *bytes, length, idx, end)

		if end > 0 {
			*list = append(*list, (*bytes)[:end])
		}
	}

	return length, idx
}

// nolint:gocritic //ignore this.
func (r *Router) Match(bytes []byte, length, index int) ([]byte, int) {
	start := index
	index = util.IndexLineEnd(bytes, length, index)

	if !r.Matches(bytes[start:index]) {
		index = util.IgnoreLineEnd(bytes, length, index)

		return nil, index
	}

	end := index

	index = util.IgnoreLineEnd(bytes, length, index)

	// append following lines
	index, end = IndexFollowingLines(r.Format, bytes, length, index, end)

	return bytes[start:end], index
}

// nolint:gocritic //ignore this.
func IndexFollowingLines(format *match.Format, bytes []byte, length, index, end int) (int, int) {
	for index < length && match.IsFollowingLine(format, bytes[index:]) {
		index = util.IndexLineEnd(bytes, length, index)

		end = index

		index = util.IgnoreLineEnd(bytes, length, index)
	}

	return index, end
}

func (r *Router) Trans(bytes ...[]byte) error {
	transfers := r.Transfers
	if len(transfers) == 0 {
		return nil
	}

	for _, t := range transfers {
		if err := t.Trans(r.Source, bytes...); err != nil {
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

func (r *Router) NextBytes() []byte {
	select {
	case <-r.Runner.C:
		return nil
	case bytes := <-r.Channel:
		if bytes == nil {
			r.Stop()

			return nil
		}

		return bytes
	case <-time.After(DurationReadNextTimeout):
		return nil
	}
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
