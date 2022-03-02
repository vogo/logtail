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

package tail

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/vogo/fwatch"
	"github.com/vogo/grunner"
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/util"
	"github.com/vogo/vogo/vos"
)

const (

	// CommandFailRetryInterval command fail retry interval.
	CommandFailRetryInterval = 10 * time.Second
)

type Server struct {
	ID            string
	lock          sync.Mutex
	Runner        *grunner.Runner
	Format        *match.Format
	Tailer        *Tailer
	workerError   chan error
	workerStarter func()
	MergingWorker *Worker
	Workers       map[string]*Worker
	Routers       map[string]*Router
}

func (s *Server) addWorker(w *Worker) {
	s.Workers[w.ID] = w
}

// NewServer Start a new server.
func NewServer(serverConfig *conf.ServerConfig) *Server {
	server := &Server{
		ID:      serverConfig.Name,
		lock:    sync.Mutex{},
		Runner:  grunner.New(),
		Routers: make(map[string]*Router, util.DefaultMapSize),
		Workers: make(map[string]*Worker, util.DefaultMapSize),
	}

	return server
}

func (s *Server) Initial(config *conf.Config, serverConfig *conf.ServerConfig) {
	var routerConfigs []*conf.RouterConfig
	if len(serverConfig.Routers) > 0 {
		routerConfigs = append(routerConfigs, config.GetRouters(serverConfig.Routers)...)
	} else {
		routerConfigs = config.AppendDefaultRouters(routerConfigs)
	}

	routerConfigs = config.AppendGlobalRouters(routerConfigs)

	for _, routerConfig := range routerConfigs {
		err := s.AddRouter(routerConfig)
		if err != nil {
			logger.Warnf("add router %s error: %v", routerConfig.Name, err)
		}
	}

	s.workerStarter = func() {
		s.MergingWorker = NewWorker(s, "", false)
		s.MergingWorker.Start()

		switch {
		case serverConfig.File != nil:
			if fwatch.IsDir(serverConfig.File.Path) {
				go s.startDirWorkers(serverConfig.File)
			} else {
				s.addWorker(StartWorker(s, util.FollowRetryTailCommand(serverConfig.File.Path), false))
			}
		case serverConfig.CommandGen != "":
			go s.startCommandGenWorkers(serverConfig.CommandGen)
		case serverConfig.Commands != "":
			commands := strings.Split(serverConfig.Commands, "\n")

			for _, cmd := range commands {
				s.addWorker(StartWorker(s, cmd, false))
			}
		case serverConfig.Command != "":
			s.addWorker(StartWorker(s, serverConfig.Command, false))
		default:
			logger.Warnf("no external stream for server %s, call server.Fire([]byte) to send data", s.ID)
		}
	}
}

// Write bytes data to default workers, which will be send to web socket clients.
func (s *Server) Write(data []byte) (int, error) {
	return s.MergingWorker.WriteToFilters(data)
}

// Fire custom generate bytes data to the first worker of the server.
func (s *Server) Fire(data []byte) error {
	for _, w := range s.Workers {
		_, err := w.Write(data)

		return err
	}

	return nil
}

// Start the server.
// First, Start all routers.
// Then, call the Start func.
func (s *Server) Start() {
	for _, r := range s.Routers {
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
			logger.Warnf("server %s close error: %+v", s.ID, err)
		}
	}()

	logger.Infof("server %s stopping", s.ID)
	s.Runner.Stop()

	s.stopWorkers()

	s.MergingWorker.Stop()

	for _, r := range s.Routers {
		r.Stop()
	}

	return nil
}

// stopWorkers stop all Workers of server, but not for the merging worker.
func (s *Server) stopWorkers() {
	for k, w := range s.Workers {
		w.Stop()

		// fix nil exception
		delete(s.Workers, k)
	}
}

// ShutdownWorker stop all workers of server, but not for the merging worker.
func (s *Server) ShutdownWorker(worker *Worker) {
	delete(s.Workers, worker.ID)

	// close worker stop chan.
	worker.Runner.Stop()

	// call worker stop.
	worker.Stop()
}

// startCommandGenWorkers Start Workers using generated commands.
// When one of Workers has error, stop all Workers,
// and generate new commands to create new Workers.
func (s *Server) startCommandGenWorkers(gen string) {
	var (
		err      error
		commands []byte
	)

	for {
		select {
		case <-s.Runner.C:
			return
		default:
			commands, err = vos.ExecShell(gen)
			if err != nil {
				logger.Errorf("server [%s] command error: %+v, command: %s", s.ID, err, gen)
			} else {
				// create a new chan everytime
				s.workerError = make(chan error)

				cmds := bytes.Split(commands, []byte{'\n'})
				for _, cmd := range cmds {
					s.addWorker(StartWorker(s, string(cmd), true))
				}

				// wait any error from one of worker
				err = <-s.workerError
				logger.Errorf("server [%s] receive worker error: %+v", s.ID, err)
				close(s.workerError)

				s.stopWorkers()
			}

			select {
			case <-s.Runner.C:
				return
			default:
				logger.Errorf("server [%s] failed, retry after 10s!", s.ID)
				time.Sleep(CommandFailRetryInterval)
			}
		}
	}
}

func (s *Server) ReceiveWorkerError(err error) {
	defer func() {
		// ignore chan closed error
		_ = recover()
	}()

	s.workerError <- err
}

// startDirWorkers Start Workers using file config.
func (s *Server) startDirWorkers(config *conf.FileConfig) {
	watcher, err := fwatch.New(config.Method, fileInactiveDeadline, fileSilenceDeadline)
	if err != nil {
		logger.Fatal(err)
	}

	matcher := func(name string) bool {
		return (config.Prefix == "" || strings.HasPrefix(name, config.Prefix)) &&
			(config.Suffix == "" || strings.HasSuffix(name, config.Suffix))
	}

	logger.Infof("server [%s] Start watch directory: %s", s.ID, config.Path)

	// start watch loop first
	go s.startDirWatchWorkers(config.Path, watcher)

	if err = watcher.WatchDir(config.Path, config.Recursive, matcher); err != nil {
		logger.Fatal(err)
	}
}

// file inactive deadline, default one hour.
const fileInactiveDeadline = time.Hour

// file inactive deadline, default one day.
const fileSilenceDeadline = time.Hour * 24

func (s *Server) startDirWatchWorkers(path string, watcher *fwatch.FileWatcher) {
	defer func() {
		_ = watcher.Stop()

		logger.Infof("server [%s] stop watch directory: %s", s.ID, path)
	}()

	fileWorkerMap := make(map[string]*Worker, util.DefaultMapSize)

	for {
		select {
		case err := <-s.workerError:
			// only log worker error
			logger.Errorf("server [%s] receive worker error: %+v", s.ID, err)
		case <-watcher.Runner.C:
			return
		case <-s.Runner.C:
			return
		case watchError := <-watcher.Errors:
			logger.Errorf("watch error: %v", watchError)
		case watchEvent := <-watcher.Events:
			switch watchEvent.Event {
			case fwatch.Create, fwatch.Write:
				logger.Infof("notify active file: %s", watchEvent.Name)

				if w, ok := fileWorkerMap[watchEvent.Name]; ok {
					logger.Infof("worker [%s] is already tailing file: %s", w.ID, watchEvent.Name)
				} else {
					// non-dynamic worker will retry self
					w := StartWorker(s, util.FollowRetryTailCommand(watchEvent.Name), false)
					w.Runner = s.Runner.NewChild()
					fileWorkerMap[watchEvent.Name] = w
					s.addWorker(w)
				}
			case fwatch.Inactive:
				logger.Infof("notify inactive file: %s", watchEvent.Name)

				if w, ok := fileWorkerMap[watchEvent.Name]; ok {
					w.Shutdown()
					delete(fileWorkerMap, watchEvent.Name)
				}
			case fwatch.Remove, fwatch.Silence:
				logger.Infof("notify remove file: %s", watchEvent.Name)

				if w, ok := fileWorkerMap[watchEvent.Name]; ok {
					w.Shutdown()
					delete(fileWorkerMap, watchEvent.Name)
				}
			default:
				logger.Warnf("unknown event: %s, %s", watchEvent.Event, watchEvent.Name)
			}
		}
	}
}

func (s *Server) AddRouter(routerConfig *conf.RouterConfig) error {
	if r, exist := s.Routers[routerConfig.Name]; exist {
		r.Stop()
	}

	router := BuildRouter(s, routerConfig)

	if err := router.Start(); err != nil {
		return err
	}

	s.Routers[routerConfig.Name] = router

	return nil
}
