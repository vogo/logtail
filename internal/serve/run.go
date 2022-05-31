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
	"runtime/debug"
	"strings"

	"github.com/vogo/fwatch"
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/util"
)

// Start the server. Call the Start func to start worker.
func (s *Server) Start(serverConfig *conf.ServerConfig) {
	s.MergingWorker = s.buildWorker("", false)
	// default no router for merging worker.
	s.MergingWorker.RouterConfigsFunc = func() []*conf.RouterConfig {
		return nil
	}
	go s.MergingWorker.StartLoop()

	switch {
	case serverConfig.File != nil:
		if fwatch.IsDir(serverConfig.File.Path) {
			go s.startDirWorkers(serverConfig.File)
		} else {
			s.AddWorker(util.FollowRetryTailCommand(serverConfig.File.Path), false)
		}
	case serverConfig.CommandGen != "":
		go s.StartCommandGenLoop(serverConfig.CommandGen)
	case serverConfig.Commands != "":
		commands := strings.Split(serverConfig.Commands, "\n")

		for _, cmd := range commands {
			s.AddWorker(cmd, false)
		}
	case serverConfig.Command != "":
		s.AddWorker(serverConfig.Command, false)
	default:
		logger.Warnf("no external stream for server %s, call server.Write([]byte) to send data", s.ID)
	}
}

// Stop the server.
func (s *Server) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	defer func() {
		if err := recover(); err != nil {
			logger.Warnf("server %s close error: %+v, stack:\n%s", s.ID, err, string(debug.Stack()))
		}
	}()

	logger.Infof("server %s stopping", s.ID)
	s.Runner.Stop()

	s.StopWorkers()

	s.MergingWorker.Stop()

	return nil
}
