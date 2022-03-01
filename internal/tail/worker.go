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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/vogo/grunner"
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/util"
)

var ErrWorkerCommandStopped = errors.New("worker command stopped")

type Worker struct {
	mu      sync.Mutex
	ID      string
	Server  *Server
	Runner  *grunner.Runner
	dynamic bool      // command generated dynamically
	command string    // command lines
	cmd     *exec.Cmd // command object
	filters map[string]*Filter
}

func (w *Worker) Write(data []byte) (int, error) {
	// copy data to avoid being update by source
	newData := make([]byte, len(data))
	copy(newData, data)

	for _, r := range w.filters {
		r.Receive(newData)
	}

	_, _ = w.Server.Write(newData)

	return len(newData), nil
}

func (w *Worker) WriteToFilters(bytes []byte) (int, error) {
	for _, r := range w.filters {
		r.Receive(bytes)
	}

	return len(bytes), nil
}

func (w *Worker) StartRouterFilter(router *Router) {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.Runner.C:
		return
	default:
		filter := NewFilter(w, router)
		w.filters[router.ID] = filter

		go func() {
			defer delete(w.filters, router.ID)
			filter.Start()
		}()
	}
}

// nolint:gosec //ignore this.
func (w *Worker) Start() {
	go func() {
		defer func() {
			w.Stop()
			logger.Infof("worker [%s] stopped", w.ID)
		}()

		if w.command == "" {
			<-w.Runner.C

			return
		}

		for {
			select {
			case <-w.Runner.C:
				return
			default:
				logger.Infof("worker [%s] command: %s", w.ID, w.command)

				w.cmd = exec.Command("/bin/sh", "-c", w.command)

				SetCmdSysProcAttr(w.cmd)

				w.cmd.Stdout = w
				w.cmd.Stderr = os.Stderr

				if err := w.cmd.Run(); err != nil {
					logger.Errorf("worker [%s] command error: %+v, command: %s", w.ID, err, w.command)

					// if the command is generated dynamic, should not restart by self, send error instead.
					if w.dynamic {
						w.Server.ReceiveWorkerError(err)

						return
					}

					select {
					case <-w.Runner.C:
						return
					default:
						logger.Errorf("worker [%s] failed, retry after 10s! command: %s", w.ID, w.command)
						time.Sleep(CommandFailRetryInterval)
					}
				}

				// if the command is generated dynamic, should not restart by self, send error instead.
				if w.dynamic {
					w.Server.ReceiveWorkerError(fmt.Errorf("%w: worker [%s]", ErrWorkerCommandStopped, w.ID))

					return
				}
			}
		}
	}()
}

// Stop will Stop the current Worker, but it may retry to Start later.
// it will not close the Tailer.stop chan.
func (w *Worker) Stop() {
	defer func() {
		if err := recover(); err != nil {
			logger.Warnf("worker [%s] close error: %+v", w.ID, err)
		}
	}()

	if w.cmd != nil {
		logger.Infof("worker [%s] command stopping: %s", w.ID, w.command)

		if err := KillCmd(w.cmd); err != nil {
			logger.Warnf("worker [%s] kill command error: %+v", w.ID, err)
		}

		w.cmd = nil
	}

	w.stopFilters()
}

// Shutdown will close the current Worker, even may close the server,
// depending on the effect scope of the Tailer.
func (w *Worker) Shutdown() {
	// let server do the worker shutdown.
	w.Server.ShutdownWorker(w)
}

func (w *Worker) stopFilters() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, filter := range w.filters {
		filter.Stop()
	}
}

func StartWorker(s *Server, command string, dynamic bool) *Worker {
	runWorker := NewWorker(s, command, dynamic)

	if len(s.Routers) > 0 {
		for _, r := range s.Routers {
			runWorker.StartRouterFilter(r)
		}
	}

	runWorker.Start()

	return runWorker
}

func NewWorker(workerServer *Server, command string, dynamic bool) *Worker {
	workerID := fmt.Sprintf("%s-%d", workerServer.ID, len(workerServer.Workers))
	if command == "" {
		workerID = fmt.Sprintf("%s-default", workerServer.ID)
	}

	return &Worker{
		mu:      sync.Mutex{},
		ID:      workerID,
		Server:  workerServer,
		Runner:  workerServer.Runner,
		command: command,
		dynamic: dynamic,
		filters: make(map[string]*Filter, util.DefaultMapSize),
	}
}
