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
	"sync"

	"github.com/vogo/gorun"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/logtail/internal/util"
	"github.com/vogo/logtail/internal/work"
)

type Server struct {
	ID                string
	lock              sync.Mutex
	Runner            *gorun.Runner
	Format            *match.Format
	TransferMatcher   trans.TransferMatcher
	RouterConfigsFunc conf.RouterConfigsFunc
	workerError       chan error
	MergingWorker     *work.Worker
	WorkerIndex       int
	Workers           map[string]*work.Worker
}

// NewRawServer StartLoop a new server.
func NewRawServer(id string) *Server {
	server := &Server{
		ID:      id,
		lock:    sync.Mutex{},
		Runner:  gorun.New(),
		Workers: make(map[string]*work.Worker, util.DefaultMapSize),
	}

	return server
}

// Write custom generate bytes data to the first worker of the server.
func (s *Server) Write(data []byte) (int, error) {
	for _, w := range s.Workers {
		_, _ = w.Write(data)
	}

	// Write bytes data to default workers, which will be send to web socket clients.
	return s.MergingWorker.WriteToRouters(data)
}
