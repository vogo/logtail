package logtail

import (
	"fmt"
	"sync"
	"time"
)

const DurationReadNextTimeout = time.Millisecond * 60

type Router struct {
	id        int64
	name      string
	lock      sync.Mutex
	once      sync.Once
	close     chan struct{}
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

func (r *Router) Start() error {
	for _, t := range r.transfers {
		if err := t.start(r); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) Stop() {
	r.once.Do(func() {
		close(r.close)
	})
}
