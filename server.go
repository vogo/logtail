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
	command     string
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
		command:     command,
		routers:     make(map[int64]*Router, 4),
		routerCount: 0,
	}
	if existsServer, ok := serverDB[id]; ok {
		_ = existsServer.Stop()
		delete(serverDB, id)
	}

	serverDB[id] = server

	return server
}

func (s *Server) Write(bytes []byte) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	message := &Message{
		ServerId: s.id,
		Data:     bytes,
	}

	if len(s.routers) == 0 {
		for _, r := range defaultRouters {
			r.receive(message)
		}
	} else {
		for _, r := range s.routers {
			r.receive(message)
		}
	}

	for _, r := range globalRouters {
		r.receive(message)
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
		if router.id == "" {
			router.id = fmt.Sprintf("%s-%d", s.id, index)
		}

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
				logger.Infof("server %s command: %s", s.id, s.command)
				s.cmd = exec.Command("/bin/sh", "-c", s.command)
				s.cmd.Stdout = s
				s.cmd.Stderr = s

				if err := s.cmd.Run(); err != nil {
					select {
					case <-s.stop:
						return
					default:
						logger.Errorf("failed to exec command, retry after 10s! error: %+v, command: %s", err, s.command)
						time.Sleep(10 * time.Second)
					}
				}
			}
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

	if s.cmd != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			logger.Warnf("server %s kill command error: %+v", s.id, err)
		}
		s.cmd = nil
	}

	s.StopRouters()

	return nil
}

func (s *Server) StopRouters() {
	for _, router := range s.routers {
		router.Stop()
	}
}
