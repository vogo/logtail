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
	"runtime/debug"

	"github.com/vogo/logger"
)

func (w *Worker) NotifyError(err error) {
	defer func() {
		// ignore chan closed error
		_ = recover()
	}()

	w.ErrorChan <- err
}

func (w *Worker) Write(data []byte) (int, error) {
	// copy data to avoid being update by source
	newData := make([]byte, len(data))
	copy(newData, data)

	for _, r := range w.Routers {
		r.Receive(newData)
	}

	if w.MergingWorker != nil {
		_, _ = w.MergingWorker.Write(newData)
	}

	return len(newData), nil
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
			logger.Warnf("worker [%s] close error: %+v, stack:\n%s", w.ID, err, string(debug.Stack()))
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
