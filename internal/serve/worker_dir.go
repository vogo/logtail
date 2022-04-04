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

package serve

import (
	"strings"
	"time"

	"github.com/vogo/fwatch"
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/util"
	"github.com/vogo/logtail/internal/work"
)

// startDirWorkers Start Workers using file config.
func (s *Server) startDirWorkers(config *conf.FileConfig) {
	watcher, err := fwatch.New(config.Method, fileInactiveDeadline, fileSilenceDeadline)
	if err != nil {
		logger.Fatal(err)
	}

	// start watch loop first
	go s.startDirWatchWorkers(config.Path, watcher)

	matcher := func(name string) bool {
		return (config.Prefix == "" || strings.HasPrefix(name, config.Prefix)) &&
			(config.Suffix == "" || strings.HasSuffix(name, config.Suffix))
	}

	if config.DirFileCountLimit > 0 {
		watcher.SetDirFileCountLimit(config.DirFileCountLimit)
	}

	logger.Infof("server [%s] StartLoop watch directory: %s", s.ID, config.Path)

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

	fileWorkerMap := make(map[string]*work.Worker, util.DefaultMapSize)

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
					w = s.AddWorker(util.FollowRetryTailCommand(watchEvent.Name), false)
					fileWorkerMap[watchEvent.Name] = w
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
