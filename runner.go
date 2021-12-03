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
		_, err := r.AddServer(serverConfig)
		if err != nil {
			logger.Errorf("add server %s error: %v", serverConfig.Name, err)
		}
	}

	return nil
}

func (r *Runner) startTransfers() error {
	for _, c := range r.Config.Transfers {
		if _, err := r.StartTransfer(c); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) AddTransfer(c *TransferConfig) error {
	if _, err := r.StartTransfer(c); err != nil {
		return err
	}

	r.Config.Transfers[c.Name] = c
	r.Config.saveToFile()

	return nil
}

// nolint:ireturn //ignore this.
func (r *Runner) StartTransfer(c *TransferConfig) (transfer.Transfer, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	t := buildTransfer(c)

	if err := t.Start(); err != nil {
		logger.Infof("transfer [%s]%s Start error: %v", c.Type, t.Name(), err)

		return nil, err
	}

	logger.Infof("transfer [%s]%s started", c.Type, t.Name())

	existTransfer, exist := r.Transfers[c.Name]

	// save or replace transfer
	r.Transfers[c.Name] = t

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

func (r *Runner) StopTransfer(name string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if existTransfer, exist := r.Transfers[name]; exist {
		if r.usingTransfer(name) {
			return fmt.Errorf("%w: %s", ErrTransferUsing, name)
		}

		err := existTransfer.Stop()
		if err != nil {
			logger.Warnf("stop transfer error: %v", err)
		}

		delete(r.Transfers, name)

		delete(r.Config.Transfers, name)
		r.Config.saveToFile()
	}

	return nil
}

func (r *Runner) usingTransfer(name string) bool {
	for _, server := range r.Servers {
		for _, router := range server.routers {
			for i := range router.transfers {
				if router.transfers[i].Name() == name {
					return true
				}
			}
		}
	}

	return false
}

// Stop the runner.
func (r *Runner) Stop() {
	r.lock.Lock()
	defer r.lock.Unlock()

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

func (r *Runner) AddRouter(config *RouterConfig) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if err := checkRouterConfig(r.Config, config); err != nil {
		return err
	}

	var err error

	if _, ok := r.Config.Routers[config.Name]; ok {
		for _, server := range r.Servers {
			for _, router := range server.routers {
				if router.Name == config.Name {
					if err = server.addRouter(config); err != nil {
						logger.Errorf("add router error: %v", err)
					}
				}
			}
		}
	}

	r.Config.Routers[config.Name] = config
	r.Config.saveToFile()

	return err
}

func (r *Runner) DeleteRouter(name string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, exist := r.Config.Routers[name]; exist {
		if r.usingRouter(name) {
			return fmt.Errorf("%w: %s", ErrRouterUsing, name)
		}

		delete(r.Config.Routers, name)
		r.Config.saveToFile()
	}

	return nil
}

func (r *Runner) usingRouter(name string) bool {
	for _, server := range r.Servers {
		for _, router := range server.routers {
			if router.Name == name {
				return true
			}
		}
	}

	return false
}

func (r *Runner) AddServer(serverConfig *ServerConfig) (*Server, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if err := checkServerConfig(r.Config, serverConfig); err != nil {
		return nil, err
	}

	server := NewServer(serverConfig)

	format := serverConfig.Format
	if format == nil {
		format = r.Config.DefaultFormat
	}

	server.format = format
	server.runner = r

	if existsServer, ok := r.Servers[server.id]; ok {
		_ = existsServer.Stop()

		delete(r.Servers, server.id)
	}

	r.Servers[server.id] = server

	server.initial(r.Config, serverConfig)

	server.Start()

	r.Config.Servers[serverConfig.Name] = serverConfig
	r.Config.saveToFile()

	return server, nil
}

func (r *Runner) DeleteServer(name string) error {
	s, exist := r.Servers[name]

	if exist {
		if err := s.Stop(); err != nil {
			return err
		}

		delete(r.Servers, name)

		delete(r.Config.Servers, name)
		r.Config.saveToFile()
	}

	return nil
}
