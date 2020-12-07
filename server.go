package logtail

import (
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
)

const DefaultServerId = "default"

var serverDBLock = sync.Mutex{}
var serverDB = make(map[string]*Server, 4)

type Server struct {
	id          string
	lock        sync.Mutex
	once        sync.Once
	stop        chan struct{}
	cmd         *exec.Cmd
	routerCount int64
	routers     map[int64]*Router
}

func NewServer(id string, command string) *Server {
	server := &Server{
		id:          id,
		lock:        sync.Mutex{},
		once:        sync.Once{},
		stop:        make(chan struct{}),
		routers:     make(map[int64]*Router, 4),
		routerCount: 0,
	}
	if existsServer, ok := serverDB[id]; ok {
		_ = existsServer.Stop()
		delete(serverDB, id)
	}

	serverDB[id] = server

	logger.Infof("server %s command: %s", id, command)

	server.cmd = exec.Command("/bin/sh", "-c", command)
	server.cmd.Stdout = server
	server.cmd.Stderr = server

	return server
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

func (s *Server) StartRouter(router *Router) {
	s.lock.Lock()
	defer s.lock.Unlock()

	select {
	case <-s.stop:
		return
	default:
		index := atomic.AddInt64(&s.routerCount, 1)
		router.id = fmt.Sprintf("%s-%d", s.id, index)

		s.routers[index] = router

		go func() {
			defer delete(s.routers, index)
			router.Start()
		}()
	}
}

func (s *Server) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()

	go func() {
		for {
			select {
			case <-s.stop:
				return
			default:
				if err := s.cmd.Run(); err != nil {
					select {
					case <-s.stop:
						return
					default:
						logger.Errorf("failed to tail file, try after 10s! error: %+v", err)
						time.Sleep(10 * time.Second)
					}
				}}
		}
	}()
}

func (s *Server) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	defer func() {
		if err := recover(); err != nil {
			logger.Warnf("server %s stop error: %+v", s.id, err)
		}
	}()

	s.once.Do(func() {
		close(s.stop)
	})

	if err := s.cmd.Process.Kill(); err != nil {
		logger.Warnf("server %s kill command error: %+v", s.id, err)
	}

	for _, t := range s.routers {
		t.Stop()
	}

	return nil
}
