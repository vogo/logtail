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
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/util"
)

func (w *Worker) NotifyError(err error) {
	defer func() {
		// ignore chan closed error
		_ = recover()
	}()

	w.ErrorChan <- err
}

// Write data to buffer, flush the buffer if it's completed.
// The buffer may not be flushed if no new data.
func (w *Worker) Write(data []byte) (int, error) {
	dataLen := len(data)
	if dataLen == 0 {
		return 0, nil
	}

	var firstLog []byte

	for len(data) > 0 {
		firstLog, data = match.SplitFirstLog(w.Format, data)

		w.buf = append(w.buf, firstLog...)

		if len(data) > 0 || firstLog[len(firstLog)-1] == '\n' {
			w.flushData(w.buf)

			// reset buffer
			w.buf = nil
		}
	}

	return dataLen, nil
}

func (w *Worker) flushData(data []byte) {
	for _, r := range w.Routers {
		r.Receive(data)
	}

	if w.MergingWorker != nil {
		w.MergingWorker.flushData(data)
	}
}

func (w *Worker) StopRouters() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, router := range w.Routers {
		router.Stop()
	}
}

// Shutdown will close the current Worker, even may close the server,
// depending on the effect scope of the Tailer.
func (w *Worker) Shutdown() {
	// close worker stop chan.
	w.Runner.Stop()

	// call worker stop.
	w.Stop()
}

// Stop will Stop the current Worker, but it may retry to StartLoop later.
// it will not close the Tailer.stop chan.
func (w *Worker) Stop() {
	defer func() {
		if err := recover(); err != nil {
			logger.Warnf("worker [%s] close error: %+v, stack:\n%s", w.ID, err, util.AllStacks())
		}
	}()

	if w.cmd != nil {
		logger.Infof("worker [%s] command stopping: %s", w.ID, w.command)

		if err := KillCmd(w.cmd); err != nil {
			logger.Warnf("worker [%s] kill command error: %+v", w.ID, err)
		}

		w.cmd = nil
	}

	w.StopRouters()
}
