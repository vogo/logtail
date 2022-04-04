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
	"bytes"
	"time"

	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/work"
	"github.com/vogo/vogo/vos"
)

// StartCommandGenLoop Start Workers using generated commands.
// When one of Workers has error, stop all Workers,
// and generate new commands to create new Workers.
func (s *Server) StartCommandGenLoop(gen string) {
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
					s.AddWorker(string(cmd), true)
				}

				// wait any error from one of worker
				err = <-s.workerError
				logger.Errorf("server [%s] receive worker error: %+v", s.ID, err)
				close(s.workerError)

				s.StopWorkers()
			}

			select {
			case <-s.Runner.C:
				return
			default:
				logger.Errorf("server [%s] failed, retry after 10s!", s.ID)
				time.Sleep(work.CommandFailRetryInterval)
			}
		}
	}
}
