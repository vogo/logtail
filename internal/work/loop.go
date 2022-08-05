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

package work

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/vogo/logger"
)

//nolint:gosec //ignore this.
func (w *Worker) StartLoop() {
	defer func() {
		w.Stop()
		logger.Infof("worker [%s] stopped", w.ID)
	}()

	routerConfigs := w.RouterConfigsFunc()

	for _, rc := range routerConfigs {
		if err := w.AddRouter(rc); err != nil {
			logger.Errorf("add router error: %v", err)
		}
	}

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
					w.NotifyError(err)

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
				w.NotifyError(fmt.Errorf("%w: worker [%s]", ErrWorkerCommandStopped, w.ID))

				return
			}
		}
	}
}
