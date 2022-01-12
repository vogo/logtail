/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logtail

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/vogo/fwatch"
	"github.com/vogo/grunner"
	"github.com/vogo/logger"
	"github.com/vogo/vogo/vos"
)

const (

	// CommandFailRetryInterval command fail retry interval.
	CommandFailRetryInterval = 10 * time.Second
)

type Server struct {
	id            string
	lock          sync.Mutex
	gorunner      *grunner.Runner
	format        *Format
	runner        *Runner
	workerError   chan error
	workerStarter func()
	MergingWorker *worker
	workers       map[string]*worker
	routers       map[string]*Router
}

func (s *Server) addWorker(w *worker) {
	s.workers[w.id] = w
}

// NewServer Start a new server.
func NewServer(serverConfig *ServerConfig) *Server {
	server := &Server{
		id:       serverConfig.Name,
		lock:     sync.Mutex{},
		gorunner: grunner.New(),
		routers:  make(map[string]*Router, defaultMapSize),
		workers:  make(map[string]*worker, defaultMapSize),
	}

	return server
}

func (s *Server) initial(config *Config, serverConfig *ServerConfig) {
	var routerConfigs []*RouterConfig
	if len(serverConfig.Routers) > 0 {
		routerConfigs = append(routerConfigs, config.GetRouters(serverConfig.Routers)...)
	} else {
		routerConfigs = config.AppendDefaultRouters(routerConfigs)
	}

	routerConfigs = config.AppendGlobalRouters(routerConfigs)

	for _, routerConfig := range routerConfigs {
		err := s.addRouter(routerConfig)
		if err != nil {
			logger.Warnf("add router %s error: %v", routerConfig.Name, err)
		}
	}

	s.workerStarter = func() {
		s.MergingWorker = newWorker(s, "", false)
		s.MergingWorker.start()

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
		case serverConfig.Command != "":
			s.addWorker(startWorker(s, serverConfig.Command, false))
		default:
			logger.Warnf("no external stream for server %s, call server.Fire([]byte) to send data", s.id)
		}
	}
}

// Write bytes data to default workers, which will be send to web socket clients.
func (s *Server) Write(data []byte) (int, error) {
	return s.MergingWorker.writeToFilters(data)
}

// Fire custom generate bytes data to the first worker of the server.
func (s *Server) Fire(data []byte) error {
	for _, w := range s.workers {
		_, err := w.Write(data)

		return err
	}

	return nil
}

// Start the server.
// First, Start all routers.
// Then, call the Start func.
func (s *Server) Start() {
	for _, r := range s.routers {
		_ = r.Start()
	}

	s.workerStarter()
}

// Stop the server.
func (s *Server) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	defer func() {
		if err := recover(); err != nil {
			logger.Warnf("server %s close error: %+v", s.id, err)
		}
	}()

	logger.Infof("server %s stopping", s.id)
	s.gorunner.Stop()

	s.stopWorkers()

	s.MergingWorker.stop()

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
func (s *Server) shutdownWorker(worker *worker) {
	delete(s.workers, worker.id)

	// close worker stop chan.
	worker.gorunner.Stop()

	// call worker stop.
	worker.stop()
}

// startCommandGenWorkers Start workers using generated commands.
// When one of workers has error, stop all workers,
// and generate new commands to create new workers.
func (s *Server) startCommandGenWorkers(gen string) {
	var (
		err      error
		commands []byte
	)

	for {
		select {
		case <-s.gorunner.C:
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
			case <-s.gorunner.C:
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

// startDirWorkers Start workers using file config.
func (s *Server) startDirWorkers(config *FileConfig) {
	watcher, err := fwatch.New(config.Method, fileInactiveDeadline, fileSilenceDeadline)
	if err != nil {
		logger.Fatal(err)
	}

	matcher := func(name string) bool {
		return (config.Prefix == "" || strings.HasPrefix(name, config.Prefix)) &&
			(config.Suffix == "" || strings.HasSuffix(name, config.Suffix))
	}

	logger.Infof("server [%s] Start watch directory: %s", s.id, config.Path)

	if err = watcher.WatchDir(config.Path, config.Recursive, matcher); err != nil {
		logger.Fatal(err)
	}

	go s.startDirWatchWorkers(config.Path, watcher)
}

// file inactive deadline, default one hour.
const fileInactiveDeadline = time.Hour

// file inactive deadline, default one day.
const fileSilenceDeadline = time.Hour * 24

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
		case <-watcher.Runner.C:
			return
		case <-s.gorunner.C:
			return
		case watchEvent := <-watcher.Events:
			switch watchEvent.Event {
			case fwatch.Create, fwatch.Write:
				logger.Infof("notify active file: %s", watchEvent.Name)

				if w, ok := fileWorkerMap[watchEvent.Name]; ok {
					logger.Infof("worker [%s] is already tailing file: %s", w.id, watchEvent.Name)
				} else {
					// non-dynamic worker will retry self
					w := startWorker(s, followRetryTailCommand(watchEvent.Name), false)
					w.gorunner = s.gorunner.NewChild()
					fileWorkerMap[watchEvent.Name] = w
					s.addWorker(w)
				}
			case fwatch.Inactive:
				logger.Infof("notify inactive file: %s", watchEvent.Name)

				if w, ok := fileWorkerMap[watchEvent.Name]; ok {
					w.shutdown()
					delete(fileWorkerMap, watchEvent.Name)
				}
			case fwatch.Remove, fwatch.Silence:
				logger.Infof("notify remove file: %s", watchEvent.Name)

				if w, ok := fileWorkerMap[watchEvent.Name]; ok {
					w.shutdown()
					delete(fileWorkerMap, watchEvent.Name)
				}
			}
		}
	}
}

func (s *Server) addRouter(routerConfig *RouterConfig) error {
	if r, exist := s.routers[routerConfig.Name]; exist {
		r.Stop()
	}

	router := buildRouter(s, routerConfig)

	if err := router.Start(); err != nil {
		return err
	}

	s.routers[routerConfig.Name] = router

	return nil
}
