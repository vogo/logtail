package logtail

import (
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
)

const DefaultServerID = "default"
const CommandFailRetryInterval = 10 * time.Second

type Server struct {
	id          string
	lock        sync.Mutex
	once        sync.Once
	stop        chan struct{}
	format      *Format
	command     string
	cmd         *exec.Cmd
	routerCount int64
	routers     map[int64]*Router
}

func NewServer(config *Config, serverConfig *ServerConfig) *Server {
	format := serverConfig.Format
	if format == nil {
		format = config.DefaultFormat
	}

	server := &Server{
		id:          serverConfig.ID,
		lock:        sync.Mutex{},
		once:        sync.Once{},
		stop:        make(chan struct{}),
		command:     serverConfig.Command,
		format:      format,
		routers:     make(map[int64]*Router, 4),
		routerCount: 0,
	}

	if existsServer, ok := serverDB[server.id]; ok {
		_ = existsServer.Stop()

		delete(serverDB, server.id)
	}

	serverDB[server.id] = server

	if len(serverConfig.Routers) > 0 {
		for _, routerConfig := range serverConfig.Routers {
			server.addRouter(buildRouter(routerConfig))
		}
	} else {
		for _, routerConfig := range config.DefaultRouters {
			server.addRouter(buildRouter(routerConfig))
		}
	}

	for _, routerConfig := range config.GlobalRouters {
		server.addRouter(buildRouter(routerConfig))
	}

	server.start()

	return server
}

func (s *Server) Write(bytes []byte) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, r := range s.routers {
		r.receive(bytes)
	}

	return len(bytes), nil
}

func (s *Server) addRouter(router *Router) {
	s.lock.Lock()
	defer s.lock.Unlock()

	router.server = s

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
			router.start()
		}()
	}
}

func (s *Server) start() {
	if s.command == "" {
		return
	}

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

				setCmdSysProcAttr(s.cmd)

				s.cmd.Stdout = s
				s.cmd.Stderr = s

				if err := s.cmd.Run(); err != nil {
					select {
					case <-s.stop:
						return
					default:
						logger.Errorf("failed to exec command, retry after 10s! error: %+v, command: %s", err, s.command)
						time.Sleep(CommandFailRetryInterval)
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
			logger.Warnf("server %s close error: %+v", s.id, err)
		}
	}()

	logger.Infof("server %s stopping", s.id)
	s.once.Do(func() {
		close(s.stop)
	})

	if s.cmd != nil {
		logger.Infof("server %s command stopping: %s", s.id, s.command)

		if err := killCmd(s.cmd); err != nil {
			logger.Warnf("server %s kill command error: %+v", s.id, err)
		}

		s.cmd = nil
	}

	s.StopRouters()

	return nil
}

func (s *Server) StopRouters() {
	for _, router := range s.routers {
		router.stop()
	}
}
