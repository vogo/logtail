package logtail

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vogo/logger"
	"github.com/vogo/vogo/vos"
)

const DefaultServerID = "default"
const CommandFailRetryInterval = 10 * time.Second

type Server struct {
	id            string
	lock          sync.Mutex
	once          sync.Once
	stop          chan struct{}
	format        *Format
	workerError   chan error
	defaultWorker *worker
	workers       []*worker
}

// NewServer start a new server.
func NewServer(config *Config, serverConfig *ServerConfig) *Server {
	format := serverConfig.Format
	if format == nil {
		format = config.DefaultFormat
	}

	server := &Server{
		id:     serverConfig.ID,
		lock:   sync.Mutex{},
		once:   sync.Once{},
		stop:   make(chan struct{}),
		format: format,
	}

	if existsServer, ok := serverDB[server.id]; ok {
		_ = existsServer.Stop()

		delete(serverDB, server.id)
	}

	serverDB[server.id] = server

	server.start(config, serverConfig)

	return server
}

// Write bytes data to default workers, which will be send to web socket clients.
func (s *Server) Write(data []byte) (int, error) {
	return s.defaultWorker.writeToRouter(data)
}

// Fire custom generate bytes data to workers.
func (s *Server) Fire(data []byte) error {
	if _, err := s.defaultWorker.writeToRouter(data); err != nil {
		return err
	}

	for _, w := range s.workers {
		if _, err := w.writeToRouter(data); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) start(config *Config, serverConfig *ServerConfig) {
	var routerConfigs []*RouterConfig
	if len(serverConfig.Routers) > 0 {
		routerConfigs = append(routerConfigs, serverConfig.Routers...)
	} else {
		routerConfigs = append(routerConfigs, config.DefaultRouters...)
	}

	routerConfigs = append(routerConfigs, config.GlobalRouters...)

	s.defaultWorker = s.startWorker(nil, "", false)

	switch {
	case serverConfig.CommandGen != "":
		s.startWorkerGen(serverConfig.CommandGen, routerConfigs)
	case serverConfig.Commands != "":
		commands := strings.Split(serverConfig.Commands, "\n")
		for _, cmd := range commands {
			s.startWorker(routerConfigs, cmd, false)
		}
	default:
		s.startWorker(routerConfigs, serverConfig.Command, false)
	}
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

	s.defaultWorker.stop()

	return nil
}

func (s *Server) stopWorkers() {
	for _, w := range s.workers {
		w.stop()
	}

	s.workers = nil
}

func (s *Server) startWorker(routerConfigs []*RouterConfig, command string, sendErrorFlag bool) *worker {
	id := fmt.Sprintf("%s-%d", s.id, len(s.workers))
	if command == "" {
		id = fmt.Sprintf("%s-default", s.id)
	}

	w := &worker{
		id:          id,
		server:      s,
		command:     command,
		dynamic:     sendErrorFlag,
		routers:     make(map[int64]*Router, 4),
		routerCount: 0,
	}

	// not add the worker into the workers list of server if no router configs.
	if len(routerConfigs) > 0 {
		s.workers = append(s.workers, w)

		for _, routerConfig := range routerConfigs {
			r := buildRouter(routerConfig)
			r.id = id + "-" + r.id
			w.addRouter(r)
		}
	}

	w.start()

	return w
}

func (s *Server) startWorkerGen(gen string, configs []*RouterConfig) {
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
						s.startWorker(configs, string(cmd), true)
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

func (s *Server) sendWorkerError(err error) {
	defer func() {
		// ignore chan closed error
		_ = recover()
	}()

	s.workerError <- err
}
