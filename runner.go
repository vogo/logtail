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

package logtail

import (
	"fmt"
	"sync"

	"github.com/vogo/logger"
	"github.com/vogo/logtail/transfer"
)

// Runner the logtail runner.
type Runner struct {
	lock      sync.Mutex
	Config    *Config
	Servers   map[string]*Server
	Transfers map[string]transfer.Transfer
}

// NewRunner new logtail runner.
func NewRunner(config *Config) (*Runner, error) {
	if err := initialCheckConfig(config); err != nil {
		return nil, err
	}

	runner := &Runner{
		lock:      sync.Mutex{},
		Config:    config,
		Servers:   make(map[string]*Server, defaultMapSize),
		Transfers: make(map[string]transfer.Transfer, defaultMapSize),
	}

	return runner, nil
}

func (r *Runner) Start() error {
	configLogLevel(r.Config.LogLevel)

	if err := r.startTransfers(); err != nil {
		return err
	}

	for _, serverConfig := range r.Config.Servers {
		r.startServer(serverConfig)
	}

	return nil
}

func (r *Runner) startTransfers() error {
	for _, c := range r.Config.transferMap {
		if _, err := r.StartTransfer(c); err != nil {
			return err
		}
	}

	return nil
}

// nolint:ireturn //ignore this.
func (r *Runner) StartTransfer(c *TransferConfig) (transfer.Transfer, error) {
	t := buildTransfer(c)

	if err := t.Start(); err != nil {
		logger.Infof("transfer [%s]%s Start error: %v", c.Type, t.Name(), err)

		return nil, err
	}

	logger.Infof("transfer [%s]%s started", c.Type, t.Name())

	existTransfer, exist := r.Transfers[c.ID]

	// save or replace transfer
	r.Transfers[c.ID] = t

	if exist {
		for _, server := range r.Servers {
			for _, router := range server.routers {
				router.lock.Lock()
				for i := range router.transfers {
					if router.transfers[i].Name() == t.Name() {
						// replace transfer
						router.transfers[i] = t
					}
				}
				router.lock.Unlock()
			}
		}

		// stop exists transfer
		_ = existTransfer.Stop()
	}

	return t, nil
}

func (r *Runner) StopTransfer(c *TransferConfig) error {
	if existTransfer, exist := r.Transfers[c.ID]; exist {
		if r.existTransfer(c) {
			return fmt.Errorf("%w: %s", ErrTransferUsing, c.ID)
		}

		err := existTransfer.Stop()
		if err != nil {
			logger.Warnf("stop transfer error: %v", err)
		}

		delete(r.Transfers, c.ID)
	}

	return nil
}

func (r *Runner) existTransfer(c *TransferConfig) bool {
	for _, server := range r.Servers {
		for _, router := range server.routers {
			for i := range router.transfers {
				if router.transfers[i].Name() == c.ID {
					return true
				}
			}
		}
	}

	return false
}

// Stop the runner.
func (r *Runner) Stop() {
	for _, s := range r.Servers {
		if err := s.Stop(); err != nil {
			logger.Errorf("server %s close error: %+v", s.id, err)
		}
	}

	for _, t := range r.Transfers {
		if err := t.Stop(); err != nil {
			logger.Errorf("transfer %s close error: %+v", t.Name(), err)
		}
	}
}

func (r *Runner) startServer(config *ServerConfig) {
	r.lock.Lock()
	defer r.lock.Unlock()

	server := NewServer(r, config)
	server.Start()
}
