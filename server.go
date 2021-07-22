package logtail

import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vogo/fwatch"
	"github.com/vogo/logger"
	"github.com/vogo/vogo/vos"
)

const (
	// DefaultServerID default server id.
	DefaultServerID = "default"

	// CommandFailRetryInterval command fail retry interval.
	CommandFailRetryInterval = 10 * time.Second
)

type Server struct {
	id            string
	lock          sync.Mutex
	stopper       *Stopper
	format        *Format
	workerError   chan error
	mergingWorker *worker
	workers       map[string]*worker
	workerStarter func()
	routersCount  int64
	routers       map[int64]*Router
}

func (s *Server) addWorker(w *worker) {
	s.workers[w.id] = w
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
		stopper: NewStopper(),
		format:  format,
		routers: make(map[int64]*Router, defaultMapSize),
		workers: make(map[string]*worker, defaultMapSize),
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
		case serverConfig.File != nil:
			if fwatch.IsDir(serverConfig.File.Path) {
				go s.startDirWorkers(serverConfig.File)
			} else {
				s.addWorker(startWorker(s, followRetryTailCommand(serverConfig.File.Path), false))
			}
		case serverConfig.CommandGen != "":
			go s.startCommandGenWorkers(serverConfig.CommandGen)
		case serverConfig.Commands != "":
			commands := strings.Split(serverConfig.Commands, "\n")

			for _, cmd := range commands {
				s.addWorker(startWorker(s, cmd, false))
			}
		default:
			s.addWorker(startWorker(s, serverConfig.Command, false))
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
	for _, w := range s.workers {
		_, err := w.Write(data)

		return err
	}

	return nil
}

// Start start server.
// First, start all routers.
// Then, call the start func.
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
	s.stopper.Stop()

	s.stopWorkers()

	s.mergingWorker.stop()

	for _, r := range s.routers {
		r.Stop()
	}

	return nil
}

// stopWorkers stop all workers of server, but not for the merging worker.
func (s *Server) stopWorkers() {
	for k, w := range s.workers {
		w.stop()

		// fix nil exception
		delete(s.workers, k)
	}
}

// stopWorkers stop all workers of server, but not for the merging worker.
func (s *Server) shutdownWorker(w *worker) {
	delete(s.workers, w.id)

	// close worker stop chan.
	w.stopper.Stop()

	// call worker stop.
	w.stop()
}

// startCommandGenWorkers start workers using generated commands.
// When one of workers has error, stop all workers,
// and generate new commands to create new workers.
func (s *Server) startCommandGenWorkers(gen string) {
	var (
		err      error
		commands []byte
	)

	for {
		select {
		case <-s.stopper.stop:
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
					s.addWorker(startWorker(s, string(cmd), true))
				}

				// wait any error from one of worker
				err = <-s.workerError
				logger.Errorf("server [%s] receive worker error: %+v", s.id, err)
				close(s.workerError)

				s.stopWorkers()
			}

			select {
			case <-s.stopper.stop:
				return
			default:
				logger.Errorf("server [%s] failed, retry after 10s!", s.id)
				time.Sleep(CommandFailRetryInterval)
			}
		}
	}
}

func (s *Server) receiveWorkerError(err error) {
	defer func() {
		// ignore chan closed error
		_ = recover()
	}()

	s.workerError <- err
}

// startDirWorkers start workers using file config.
func (s *Server) startDirWorkers(config *FileConfig) {
	watcher, err := fwatch.NewFileWatcher(config.Path, config.Recursive, config.Method, fileInactiveDeadline,
		func(name string) bool {
			return (config.Prefix == "" || strings.HasPrefix(name, config.Prefix)) &&
				(config.Suffix == "" || strings.HasSuffix(name, config.Suffix))
		})
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("server [%s] start watch directory: %s", s.id, config.Path)

	go s.startDirWatchWorkers(config.Path, watcher)

	if err = watcher.Start(); err != nil {
		logger.Fatal(err)
	}
}

// file inactive deadline, default one hour.
const fileInactiveDeadline = time.Hour

func (s *Server) startDirWatchWorkers(path string, watcher *fwatch.FileWatcher) {
	defer func() {
		_ = watcher.Stop()

		logger.Infof("server [%s] stop watch directory: %s", s.id, path)
	}()

	fileWorkerMap := make(map[string]*worker, defaultMapSize)

	for {
		select {
		case err := <-s.workerError:
			// only log worker error
			logger.Errorf("server [%s] receive worker error: %+v", s.id, err)
		case <-watcher.Done:
			return
		case <-s.stopper.stop:
			return
		case f := <-watcher.ActiveChan:
			logger.Infof("notify active file: %s", f.Name)

			if w, ok := fileWorkerMap[f.Name]; ok {
				logger.Infof("worker [%s] is already tailing file: %s", w.id, f.Name)
			} else {
				// non-dynamic worker will retry self
				w := startWorker(s, followRetryTailCommand(f.Name), false)
				w.stopper = FromStopper(s.stopper)
				fileWorkerMap[f.Name] = w
				s.addWorker(w)
			}
		case f := <-watcher.InactiveChan:
			logger.Infof("notify inactive file: %s", f.Name)

			if w, ok := fileWorkerMap[f.Name]; ok {
				w.shutdown()
				delete(fileWorkerMap, f.Name)
			}
		case name := <-watcher.RemoveChan:
			logger.Infof("notify remove file: %s", name)

			if w, ok := fileWorkerMap[name]; ok {
				w.shutdown()
				delete(fileWorkerMap, name)
			}
		}
	}
}
