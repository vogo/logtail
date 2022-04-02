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

package tail

import (
	"fmt"
	"sync"

	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/logtail/internal/util"
)

// DefaultTailer the default tailer.
// nolint:gochecknoglobals // ignore this
var DefaultTailer *Tailer

// Tailer the logtail tailer.
type Tailer struct {
	lock      sync.Mutex
	Config    *conf.Config
	Servers   map[string]*Server
	Transfers map[string]trans.Transfer
}

// NewTailer new logtail tailer.
func NewTailer(config *conf.Config) (*Tailer, error) {
	if err := conf.InitialCheckConfig(config); err != nil {
		return nil, err
	}

	tailer := &Tailer{
		lock:      sync.Mutex{},
		Config:    config,
		Servers:   make(map[string]*Server, util.DefaultMapSize),
		Transfers: make(map[string]trans.Transfer, util.DefaultMapSize),
	}

	return tailer, nil
}

func (t *Tailer) Start() error {
	conf.ConfigLogLevel(t.Config.LogLevel)

	if err := t.startTransfers(); err != nil {
		return err
	}

	for _, serverConfig := range t.Config.Servers {
		_, err := t.AddServer(serverConfig)
		if err != nil {
			logger.Errorf("add server %s error: %v", serverConfig.Name, err)
		}
	}

	return nil
}

func (t *Tailer) startTransfers() error {
	for _, c := range t.Config.Transfers {
		if _, err := t.StartTransfer(c); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tailer) AddTransfer(c *conf.TransferConfig) error {
	if _, err := t.StartTransfer(c); err != nil {
		return err
	}

	t.Config.Transfers[c.Name] = c
	t.Config.SaveToFile()

	return nil
}

// nolint:ireturn //ignore this.
func (t *Tailer) StartTransfer(transferConfig *conf.TransferConfig) (trans.Transfer, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	runTransfer := BuildTransfer(transferConfig)

	if err := runTransfer.Start(); err != nil {
		logger.Infof("transfer [%s]%s Start error: %v", transferConfig.Type, runTransfer.Name(), err)

		return nil, err
	}

	logger.Infof("transfer [%s]%s started", transferConfig.Type, runTransfer.Name())

	existTransfer, exist := t.Transfers[transferConfig.Name]

	// save or replace transfer
	t.Transfers[transferConfig.Name] = runTransfer

	if exist {
		for _, server := range t.Servers {
			for _, worker := range server.Workers {
				for _, router := range worker.Routers {
					router.Lock.Lock()
					for i := range router.Transfers {
						if router.Transfers[i].Name() == runTransfer.Name() {
							// replace transfer
							router.Transfers[i] = runTransfer
						}
					}
					router.Lock.Unlock()
				}
			}
		}

		// stop exists transfer
		_ = existTransfer.Stop()
	}

	return runTransfer, nil
}

func (t *Tailer) StopTransfer(name string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if existTransfer, exist := t.Transfers[name]; exist {
		if t.isTransferUsing(name) {
			return fmt.Errorf("%w: %s", conf.ErrTransferUsing, name)
		}

		err := existTransfer.Stop()
		if err != nil {
			logger.Warnf("stop transfer error: %v", err)
		}

		delete(t.Transfers, name)

		delete(t.Config.Transfers, name)
		t.Config.SaveToFile()
	}

	return nil
}

func (t *Tailer) isTransferUsing(name string) bool {
	for _, router := range t.Config.Routers {
		for i := range router.Transfers {
			if router.Transfers[i] == name {
				return true
			}
		}
	}

	return false
}

// Stop the runner.
func (t *Tailer) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, s := range t.Servers {
		if err := s.Stop(); err != nil {
			logger.Errorf("server %s close error: %+v", s.ID, err)
		}
	}

	for _, t := range t.Transfers {
		if err := t.Stop(); err != nil {
			logger.Errorf("transfer %s close error: %+v", t.Name(), err)
		}
	}
}

func (t *Tailer) AddRouter(config *conf.RouterConfig) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := conf.CheckRouterConfig(t.Config, config); err != nil {
		return err
	}

	var err error

	if _, ok := t.Config.Routers[config.Name]; ok {
		for _, server := range t.Servers {
			for _, worker := range server.Workers {
				for _, router := range worker.Routers {
					if router.Name == config.Name {
						if err = worker.AddRouter(config); err != nil {
							logger.Errorf("add Routers error: %v", err)
						}
					}
				}
			}
		}
	}

	t.Config.Routers[config.Name] = config
	t.Config.SaveToFile()

	return err
}

func (t *Tailer) DeleteRouter(name string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if _, exist := t.Config.Routers[name]; exist {
		if t.isRouterUsing(name) {
			return fmt.Errorf("%w: %s", conf.ErrRouterUsing, name)
		}

		delete(t.Config.Routers, name)
		t.Config.SaveToFile()
	}

	return nil
}

func (t *Tailer) isRouterUsing(name string) bool {
	for _, server := range t.Config.Servers {
		for _, router := range server.Routers {
			if router == name {
				return true
			}
		}
	}

	return false
}

func (t *Tailer) AddServer(serverConfig *conf.ServerConfig) (*Server, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := conf.CheckServerConfig(t.Config, serverConfig); err != nil {
		return nil, err
	}

	server := NewServer(serverConfig)

	format := serverConfig.Format
	if format == nil {
		format = t.Config.DefaultFormat
	}

	server.Format = format
	server.TransferMatcher = func(ids []string) []trans.Transfer {
		transfers := make([]trans.Transfer, 0, len(ids))

		for _, id := range ids {
			existTransfer, ok := t.Transfers[id]
			if !ok {
				logger.Errorf("transfer not exists: %s", id)

				continue
			}

			transfers = append(transfers, existTransfer)
		}

		return transfers
	}

	if existsServer, ok := t.Servers[server.ID]; ok {
		_ = existsServer.Stop()

		delete(t.Servers, server.ID)
	}

	t.Servers[server.ID] = server

	server.Initial(t.Config, serverConfig)

	server.Start()

	t.Config.Servers[serverConfig.Name] = serverConfig
	t.Config.SaveToFile()

	return server, nil
}

func (t *Tailer) DeleteServer(name string) error {
	s, exist := t.Servers[name]

	if exist {
		if err := s.Stop(); err != nil {
			return err
		}

		delete(t.Servers, name)

		delete(t.Config.Servers, name)
		t.Config.SaveToFile()
	}

	return nil
}
