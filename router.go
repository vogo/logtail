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
	server    *Server
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

	bytes = indexToLineStart(r.server.format, bytes)

	var (
		list  [][]byte
		match []byte
		end   int
	)

	i := 0
	l := len(bytes)
	followFlag := false

	for i < l {
		if followFlag {
			// append following lines
			i, end = indexFollowingLines(r.server.format, bytes, l, i, 0)
			if i > 0 {
				list = append(list, bytes[:end])

				if i >= l {
					bytes = r.nextBytes()
					if len(bytes) > 0 {
						i = 0
						l = len(bytes)
						followFlag = true

						continue
					}
				}
			}

			if err := r.Trans(list...); err != nil {
				return err
			}

			list = nil
			followFlag = false

			continue
		}

		match, i = r.Match(bytes, l, i)

		if len(match) > 0 {
			list = append(list, match)

			if i >= l {
				bytes = r.nextBytes()
				if len(bytes) > 0 {
					i = 0
					l = len(bytes)
					followFlag = true

					continue
				}
			}

			if err := r.Trans(list...); err != nil {
				return err
			}

			list = nil
			followFlag = false
		}
	}

	return nil
}

func (r *Router) Match(bytes []byte, l, i int) (matchBytes []byte, index int) {
	start := i
	i = indexLineEnd(bytes, l, i)

	if !r.matches(bytes[start:i]) {
		i = ignoreLineEnd(bytes, l, i)
		return nil, i
	}

	end := i

	i = ignoreLineEnd(bytes, l, i)

	// append following lines
	i, end = indexFollowingLines(r.server.format, bytes, l, i, end)

	return bytes[start:end], i
}

func indexFollowingLines(format *Format, bytes []byte, l, i, end int) (index, newEnd int) {
	for i < l && isFollowingLine(format, bytes[i:]) {
		i = indexLineEnd(bytes, l, i)

		end = i

		i = ignoreLineEnd(bytes, l, i)
	}

	return i, end
}

func (r *Router) Trans(bytes ...[]byte) error {
	transfers := r.transfers
	if len(transfers) == 0 {
		return nil
	}

	for _, t := range transfers {
		if err := t.Trans(r.server.id, bytes...); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) stop() {
	r.once.Do(func() {
		logger.Infof("router %s stopping", r.id)
		close(r.close)
		close(r.channel)
	})
}

func (r *Router) start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("router %s error: %+v", r.id, err)
		}

		logger.Infof("router %s stopped", r.id)
	}()

	logger.Infof("router %s start", r.id)

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
				logger.Warnf("router %s route error: %+v", r.id, err)
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
