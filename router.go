package logtail

import (
	"sync"
	"time"

	"github.com/vogo/logger"
)

const DurationNextBytesTimeout = time.Millisecond * 4

type Router struct {
	id        string
	channel   Channel
	lock      sync.Mutex
	once      sync.Once
	close     chan struct{}
	worker    *worker
	matchers  []Matcher
	transfers []Transfer
}

func newRouter(id string, matchers []Matcher, transfers []Transfer) *Router {
	t := &Router{
		id:        id,
		channel:   make(chan []byte, DefaultChannelBufferSize),
		lock:      sync.Mutex{},
		once:      sync.Once{},
		close:     make(chan struct{}),
		matchers:  matchers,
		transfers: transfers,
	}

	return t
}

func (r *Router) SetMatchers(matchers []Matcher) {
	r.matchers = matchers
}

func (r *Router) Route(bytes []byte) error {
	if len(r.matchers) == 0 {
		return r.Trans(bytes)
	}

	bytes = indexToLineStart(r.worker.server.format, bytes)

	var (
		list  [][]byte
		match []byte
	)

	idx := 0
	length := len(bytes)

	for idx < length {
		match = r.Match(bytes, &length, &idx)

		if len(match) > 0 {
			list = append(list, match)

			for length > 0 && idx >= length {
				r.readMoreFollowingLines(&list, &bytes, &length, &idx)
			}

			if err := r.Trans(list...); err != nil {
				return err
			}

			list = nil
		}
	}

	return nil
}

func (r *Router) readMoreFollowingLines(list *[][]byte, bytes *[]byte, length, idx *int) {
	*bytes = r.nextBytes()
	*idx = 0
	*length = len(*bytes)

	if *length > 0 {
		var end int
		// append following lines
		indexFollowingLines(r.worker.server.format, *bytes, length, idx, &end)

		if end > 0 {
			*list = append(*list, (*bytes)[:end])
		}
	}
}

func (r *Router) Match(bytes []byte, length, index *int) []byte {
	start := *index
	indexLineEnd(bytes, length, index)

	if !r.matches(bytes[start:*index]) {
		ignoreLineEnd(bytes, length, index)
		return nil
	}

	end := *index

	ignoreLineEnd(bytes, length, index)

	// append following lines
	indexFollowingLines(r.worker.server.format, bytes, length, index, &end)

	return bytes[start:end]
}

func indexFollowingLines(format *Format, bytes []byte, length, index, end *int) {
	for *index < *length && isFollowingLine(format, bytes[*index:]) {
		indexLineEnd(bytes, length, index)

		*end = *index

		ignoreLineEnd(bytes, length, index)
	}
}

func (r *Router) Trans(bytes ...[]byte) error {
	transfers := r.transfers
	if len(transfers) == 0 {
		return nil
	}

	for _, t := range transfers {
		if err := t.Trans(r.worker.server.id, bytes...); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) stop() {
	r.once.Do(func() {
		logger.Infof("worker [%s] router [%s] stopping", r.worker.id, r.id)
		close(r.close)
		close(r.channel)
	})
}

func (r *Router) start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("worker [%s] router [%s] error: %+v", r.worker.id, r.id, err)
		}

		logger.Infof("worker [%s] router [%s] stopped", r.worker.id, r.id)
	}()

	logger.Infof("worker [%s] router [%s] start", r.worker.id, r.id)

	for {
		select {
		case <-r.close:
			return
		case bytes := <-r.channel:
			if bytes == nil {
				r.stop()
				return
			}

			if err := r.Route(bytes); err != nil {
				logger.Warnf("worker [%s] router [%s] route error: %+v", r.worker.id, r.id, err)
				r.stop()
			}
		}
	}
}

func (r *Router) nextBytes() []byte {
	select {
	case <-r.close:
		return nil
	case <-time.After(DurationNextBytesTimeout):
		return nil
	case bytes := <-r.channel:
		if bytes == nil {
			r.stop()
			return nil
		}

		return bytes
	}
}

func (r *Router) receive(message []byte) {
	defer func() {
		_ = recover()
	}()

	select {
	case <-r.close:
		return
	case r.channel <- message:
	default:
	}
}

func (r *Router) matches(bytes []byte) bool {
	for _, m := range r.matchers {
		if !m.Match(bytes) {
			return false
		}
	}

	return true
}
