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
	"errors"
	"os/exec"
	"sync"
	"time"

	"github.com/vogo/grunner"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/route"
	"github.com/vogo/logtail/internal/trans"
)

const (
	// CommandFailRetryInterval command fail retry interval.
	CommandFailRetryInterval = 10 * time.Second
)

var ErrWorkerCommandStopped = errors.New("worker command stopped")

type Worker struct {
	mu      sync.Mutex
	dynamic bool      // command generated dynamically
	command string    // command lines
	cmd     *exec.Cmd // command object

	ID                string
	Source            string
	Runner            *grunner.Runner
	ErrorChan         chan error
	Routers           map[string]*route.Router
	MergingWorker     *Worker
	TransfersFunc     trans.TransferMatcher
	RouterConfigsFunc conf.RouterConfigsFunc
}

func NewRawWorker(workerID, command string, dynamic bool) *Worker {
	return &Worker{
		mu:      sync.Mutex{},
		ID:      workerID,
		command: command,
		dynamic: dynamic,
		Routers: make(map[string]*route.Router),
	}
}
