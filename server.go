package logtail

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
	"github.com/vogo/vogo/vos"
)

const (
	DefaultServerID          = "default"
	CommandFailRetryInterval = 10 * time.Second
)

type Server struct {
	id            string
	lock          sync.Mutex
	once          sync.Once
	stop          chan struct{}
	format        *Format
	workerError   chan error
	mergingWorker *worker
	workers       []*worker
	workerStarter func()
	routersCount  int64
	routers       map[int64]*Router
}

// NewServer start a new server.
func NewServer(config *Config, serverConfig *ServerConfig) *Server {
	format := serverConfig.Format
	if format == nil {
		format = config.DefaultFormat
	}

	server := &Server{
		id:      serverConfig.ID,
		lock:    sync.Mutex{},
		once:    sync.Once{},
		stop:    make(chan struct{}),
		format:  format,
		routers: make(map[int64]*Router, 4),
	}

	if existsServer, ok := serverDB[server.id]; ok {
		_ = existsServer.Stop()

		delete(serverDB, server.id)
	}

	serverDB[server.id] = server

	server.initial(config, serverConfig)

	return server
}

func (s *Server) initial(config *Config, serverConfig *ServerConfig) {
	var routerConfigs []*RouterConfig
	if len(serverConfig.Routers) > 0 {
		routerConfigs = append(routerConfigs, serverConfig.Routers...)
	} else {
		routerConfigs = append(routerConfigs, config.DefaultRouters...)
	}

	routerConfigs = append(routerConfigs, config.GlobalRouters...)

	// not add the worker into the workers list of server if no router configs.
	routerCount := len(routerConfigs)
	if routerCount > 0 {
		for _, routerConfig := range routerConfigs {
			r := buildRouter(s, routerConfig)
			if routerCount == 1 {
				r.name = s.id
			}

			s.routers[r.id] = r
		}
	}

	s.workerStarter = func() {
		s.mergingWorker = newWorker(s, "", false)
		s.mergingWorker.start()

		switch {
		case serverConfig.CommandGen != "":
			s.startDynamicWorkers(serverConfig.CommandGen)
		case serverConfig.Commands != "":
			commands := strings.Split(serverConfig.Commands, "\n")

			for _, cmd := range commands {
				s.workers = append(s.workers, startWorker(s, cmd, false))
			}
		default:
			s.workers = append(s.workers, startWorker(s, serverConfig.Command, false))
		}
	}
}

func (s *Server) nextRouterID() int64 {
	return atomic.AddInt64(&s.routersCount, 1)
}

// Write bytes data to default workers, which will be send to web socket clients.
func (s *Server) Write(data []byte) (int, error) {
	return s.mergingWorker.writeToFilters(data)
}

// Fire custom generate bytes data to the first worker of the server.
func (s *Server) Fire(data []byte) error {
	_, err := s.workers[0].Write(data)

	return err
}

func (s *Server) Start() {
	for _, r := range s.routers {
		_ = r.Start()
	}

	s.workerStarter()
}

// Stop stop server.
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

	s.stopWorkers()

	s.mergingWorker.stop()

	for _, r := range s.routers {
		r.Stop()
	}

	return nil
}

func (s *Server) stopWorkers() {
	for _, w := range s.workers {
		w.stop()
	}

	// fix nil exception
	s.workers = []*worker{}

	//s.workers = nil
}

func (s *Server) startDynamicWorkers(gen string) {
	go func() {
		var (
			err      error
			commands []byte
		)

		for {
			select {
			case <-s.stop:
				return
			default:
				commands, err = vos.ExecShell(gen)
				if err != nil {
					logger.Errorf("server [%s] command error: %+v, command: %s", s.id, err, gen)
				} else {
					// create a new chan everytime
					s.workerError = make(chan error)

					cmds := bytes.Split(commands, []byte{'\n'})
					for _, cmd := range cmds {
						s.workers = append(s.workers, startWorker(s, string(cmd), true))
					}

					// wait any error from one of worker
					err = <-s.workerError
					logger.Errorf("server [%s] receive worker error: %+v", s.id, err)
					close(s.workerError)

					s.stopWorkers()
				}

				select {
				case <-s.stop:
					return
				default:
					logger.Errorf("server [%s] failed, retry after 10s!", s.id)
					time.Sleep(CommandFailRetryInterval)
				}
			}
		}
	}()
}

func (s *Server) receiveWorkerError(err error) {
	defer func() {
		// ignore chan closed error
		_ = recover()
	}()

	s.workerError <- err
}
