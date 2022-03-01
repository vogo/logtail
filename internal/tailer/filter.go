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
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/util"
)

type Filter struct {
	id       string
	channel  Channel
	lock     sync.Mutex
	gorunner *grunner.Runner
	worker   *Worker
	router   *Router
}

func NewFilter(worker *Worker, router *Router) *Filter {
	routerFilter := &Filter{
		id:       fmt.Sprintf("%s-%s", worker.ID, router.ID),
		channel:  make(chan []byte, DefaultChannelBufferSize),
		lock:     sync.Mutex{},
		gorunner: worker.Gorunner.NewChild(),
		worker:   worker,
		router:   router,
	}

	return routerFilter
}

func (f *Filter) Route(bytes []byte) error {
	if len(f.router.Matchers) == 0 {
		return f.Trans(bytes)
	}

	bytes = match.IndexToLineStart(f.worker.Server.Format, bytes)

	var (
		list    [][]byte
		matches []byte
	)

	idx := 0
	length := len(bytes)

	for idx < length {
		matches = f.Match(bytes, &length, &idx)

		if len(matches) > 0 {
			list = append(list, matches)

			for length > 0 && idx >= length {
				f.readMoreFollowingLines(&list, &bytes, &length, &idx)
			}

			if err := f.Trans(list...); err != nil {
				return err
			}

			list = nil
		}
	}

	return nil
}

func (f *Filter) readMoreFollowingLines(list *[][]byte, bytes *[]byte, length, idx *int) {
	*bytes = f.nextBytes()
	*idx = 0
	*length = len(*bytes)

	if *length > 0 {
		var end int
		// append following lines
		indexFollowingLines(f.worker.Server.Format, *bytes, length, idx, &end)

		if end > 0 {
			*list = append(*list, (*bytes)[:end])
		}
	}
}

func (f *Filter) Match(bytes []byte, length, index *int) []byte {
	start := *index
	util.IndexLineEnd(bytes, length, index)

	if !f.matches(bytes[start:*index]) {
		util.IgnoreLineEnd(bytes, length, index)

		return nil
	}

	end := *index

	util.IgnoreLineEnd(bytes, length, index)

	// append following lines
	indexFollowingLines(f.worker.Server.Format, bytes, length, index, &end)

	return bytes[start:end]
}

func indexFollowingLines(format *match.Format, bytes []byte, length, index, end *int) {
	for *index < *length && IsFollowingLine(format, bytes[*index:]) {
		util.IndexLineEnd(bytes, length, index)

		*end = *index

		util.IgnoreLineEnd(bytes, length, index)
	}
}

func IsFollowingLine(format *match.Format, bytes []byte) bool {
	if format == nil {
		format = DefaultRunner.Config.DefaultFormat
	}

	if format != nil {
		return !format.PrefixMatch(bytes)
	}

	return bytes[0] == ' ' || bytes[0] == '\t'
}

func (f *Filter) Trans(bytes ...[]byte) error {
	transfers := f.router.Transfers
	if len(transfers) == 0 {
		return nil
	}

	for _, t := range transfers {
		if err := t.Trans(f.worker.Server.ID, bytes...); err != nil {
			return err
		}
	}

	return nil
}

func (f *Filter) Stop() {
	f.gorunner.StopWith(func() {
		logger.Infof("filter [%s] stopping", f.id)
		close(f.channel)
	})
}

func (f *Filter) Start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("filter [%s] error: %+v", f.id, err)
		}

		logger.Infof("filter [%s] stopped", f.id)
	}()

	logger.Infof("filter [%s] Start", f.id)

	for {
		select {
		case <-f.gorunner.C:
			return
		case data := <-f.channel:
			if data == nil {
				f.Stop()

				return
			}

			if err := f.Route(data); err != nil {
				logger.Warnf("filter [%s] route error: %+v", f.id, err)
				f.Stop()
			}
		}
	}
}

func (f *Filter) nextBytes() []byte {
	select {
	case <-f.gorunner.C:
		return nil
	case bytes := <-f.channel:
		if bytes == nil {
			f.Stop()

			return nil
		}

		return bytes
	case <-time.After(DurationReadNextTimeout):
		return nil
	}
}

func (f *Filter) Receive(data []byte) {
	defer func() {
		_ = recover()
	}()

	select {
	case <-f.gorunner.C:
		return
	case f.channel <- data:
	default:
	}
}

func (f *Filter) matches(bytes []byte) bool {
	for _, m := range f.router.Matchers {
		if !m.Match(bytes) {
			return false
		}
	}

	return true
}
