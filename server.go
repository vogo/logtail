package logtail

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const DefaultServerId = "default"

var serverDBLock = sync.Mutex{}
var serverDB = make(map[string]*Server, 4)

type Server struct {
	id          string
	lock        sync.Mutex
	command     string
	routerCount int64
	routers     map[int64]*Router
}

func (s *Server) AddRouter(router *Router) {
	s.lock.Lock()
	defer s.lock.Unlock()

	index := atomic.AddInt64(&s.routerCount, 1)
	router.id = fmt.Sprintf("%s-%d", s.id, index)

	s.routers[index] = router
	go func() {
		defer delete(s.routers, index)
		router.Start()
	}()
}

func (s *Server) Write(bytes []byte) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.routers) == 0 {
		return len(bytes), nil
	}

	for _, t := range s.routers {
		select {
		case t.channel <- bytes:
		default:
		}
	}

	return len(bytes), nil
}
