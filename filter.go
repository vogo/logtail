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

	"github.com/vogo/logger"
)

type Filter struct {
	id      string
	channel Channel
	lock    sync.Mutex
	once    sync.Once
	close   chan struct{}
	worker  *worker
	router  *Router
}

func newFilter(worker *worker, router *Router) *Filter {
	f := &Filter{
		id:      fmt.Sprintf("%s-%d", worker.id, router.id),
		channel: make(chan []byte, DefaultChannelBufferSize),
		lock:    sync.Mutex{},
		once:    sync.Once{},
		close:   make(chan struct{}),
		worker:  worker,
		router:  router,
	}

	return f
}

func (f *Filter) Route(bytes []byte) error {
	if len(f.router.matchers) == 0 {
		return f.Trans(bytes)
	}

	bytes = indexToLineStart(f.worker.server.format, bytes)

	var (
		list  [][]byte
		match []byte
	)

	idx := 0
	length := len(bytes)

	for idx < length {
		match = f.Match(bytes, &length, &idx)

		if len(match) > 0 {
			list = append(list, match)

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
		indexFollowingLines(f.worker.server.format, *bytes, length, idx, &end)

		if end > 0 {
			*list = append(*list, (*bytes)[:end])
		}
	}
}

func (f *Filter) Match(bytes []byte, length, index *int) []byte {
	start := *index
	indexLineEnd(bytes, length, index)

	if !f.matches(bytes[start:*index]) {
		ignoreLineEnd(bytes, length, index)

		return nil
	}

	end := *index

	ignoreLineEnd(bytes, length, index)

	// append following lines
	indexFollowingLines(f.worker.server.format, bytes, length, index, &end)

	return bytes[start:end]
}

func indexFollowingLines(format *Format, bytes []byte, length, index, end *int) {
	for *index < *length && isFollowingLine(format, bytes[*index:]) {
		indexLineEnd(bytes, length, index)

		*end = *index

		ignoreLineEnd(bytes, length, index)
	}
}

func (f *Filter) Trans(bytes ...[]byte) error {
	transfers := f.router.transfers
	if len(transfers) == 0 {
		return nil
	}

	for _, t := range transfers {
		if err := t.Trans(f.worker.server.id, bytes...); err != nil {
			return err
		}
	}

	return nil
}

func (f *Filter) stop() {
	f.once.Do(func() {
		logger.Infof("filter [%s] stopping", f.id)
		close(f.close)
		close(f.channel)
	})
}

func (f *Filter) start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("filter [%s] error: %+v", f.id, err)
		}

		logger.Infof("filter [%s] stopped", f.id)
	}()

	logger.Infof("filter [%s] start", f.id)

	for {
		select {
		case <-f.close:
			return
		case data := <-f.channel:
			if data == nil {
				f.stop()

				return
			}

			if err := f.Route(data); err != nil {
				logger.Warnf("filter [%s] route error: %+v", f.id, err)
				f.stop()
			}
		}
	}
}

func (f *Filter) nextBytes() []byte {
	select {
	case <-f.close:
		return nil
	case bytes := <-f.channel:
		if bytes == nil {
			f.stop()

			return nil
		}

		return bytes
	case <-time.After(DurationReadNextTimeout):
		return nil
	}
}

func (f *Filter) receive(data []byte) {
	defer func() {
		_ = recover()
	}()

	select {
	case <-f.close:
		return
	case f.channel <- data:
	default:
	}
}

func (f *Filter) matches(bytes []byte) bool {
	for _, m := range f.router.matchers {
		if !m.Match(bytes) {
			return false
		}
	}

	return true
}
