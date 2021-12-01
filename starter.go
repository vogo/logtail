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

	"github.com/vogo/logger"
)

// StartLogtail start config servers.
func StartLogtail(config *Config) {
	defaultFormat = config.DefaultFormat

	startTransfers(config.transferMap)

	for _, serverConfig := range config.Servers {
		startServer(config, serverConfig)
	}
}

func startTransfers(transferMap map[string]*TransferConfig) {
	for _, c := range transferMap {
		if _, err := startTransfer(c); err != nil {
			panic(err)
		}
	}
}

// nolint:ireturn //ignore return interface
func startTransfer(c *TransferConfig) (Transfer, error) {
	if t, ok := transferDB[c.ID]; ok {
		return t, nil
	}

	if c.isRef() {
		return nil, fmt.Errorf("%w: %s", ErrTransferNotExist, c.ID)
	}

	t := buildTransfer(c)

	if err := t.start(); err != nil {
		logger.Infof("transfer [%s]%s start error: %v", c.Type, t.ID(), err)

		return nil, err
	}

	logger.Infof("transfer [%s]%s started", c.Type, t.ID())

	existTransfer, exist := transferDB[c.ID]

	// save or replace transfer
	transferDB[c.ID] = t

	if exist {
		for _, server := range serverDB {
			for _, router := range server.routers {
				router.lock.Lock()
				for i := range router.transfers {
					if router.transfers[i].ID() == t.ID() {
						// replace transfer
						router.transfers[i] = t
					}
				}
				router.lock.Unlock()
			}
		}

		// stop exists transfer
		_ = existTransfer.stop()
	}

	return t, nil
}

// StopLogtail stop servers.
func StopLogtail() {
	for _, s := range serverDB {
		if err := s.Stop(); err != nil {
			logger.Errorf("server %s close error: %+v", s.id, err)
		}
	}
}

func startServer(c *Config, config *ServerConfig) {
	serverDBLock.Lock()
	defer serverDBLock.Unlock()

	server := NewServer(c, config)
	server.Start()
}
