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
	"fmt"

	"github.com/vogo/logtail/internal/work"
)

func (s *Server) AddWorker(command string, dynamic bool) *work.Worker {
	worker := s.StartWorker(command, dynamic)

	s.Workers[worker.ID] = worker

	return worker
}

func (s *Server) StartWorker(command string, dynamic bool) *work.Worker {
	s.WorkerIndex++
	workerID := fmt.Sprintf("%s-%d", s.ID, s.WorkerIndex)

	worker := work.NewRawWorker(workerID, command, dynamic)

	worker.Source = s.ID
	worker.Runner = s.Runner.NewChild()
	worker.ErrorChan = s.workerError
	worker.TransfersFunc = s.TransferMatcher
	worker.RouterConfigsFunc = s.RouterConfigsFunc
	worker.MergingWorker = s.MergingWorker

	go worker.StartLoop()

	return worker
}

// StopWorkers stop all Workers of server, but not for the merging worker.
func (s *Server) StopWorkers() {
	for k, w := range s.Workers {
		w.Stop()

		// fix nil exception
		delete(s.Workers, k)
	}
}

// ShutdownWorker stop all workers of server, but not for the merging worker.
func (s *Server) ShutdownWorker(worker *work.Worker) {
	delete(s.Workers, worker.ID)
}
