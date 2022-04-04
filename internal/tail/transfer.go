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

	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/trans"
)

func (t *Tailer) StartTransfers() error {
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
		logger.Infof("transfer [%s]%s StartLoop error: %v", transferConfig.Type, runTransfer.Name(), err)

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

func (t *Tailer) RemoveTransfer(name string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if existTransfer, exist := t.Transfers[name]; exist {
		if t.IsTransferUsing(name) {
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

func (t *Tailer) IsTransferUsing(name string) bool {
	for _, router := range t.Config.Routers {
		for i := range router.Transfers {
			if router.Transfers[i] == name {
				return true
			}
		}
	}

	return false
}

func buildTransferMatcher(t *Tailer) func(ids []string) []trans.Transfer {
	return func(ids []string) []trans.Transfer {
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
}

// nolint:ireturn // return diff transfer implementation.
func BuildTransfer(config *conf.TransferConfig) trans.Transfer {
	switch config.Type {
	case trans.TypeWebhook:
		return trans.NewWebhookTransfer(config.Name, config.URL, config.Prefix)
	case trans.TypeDing:
		return trans.NewDingTransfer(config.Name, config.URL, config.Prefix)
	case trans.TypeLark:
		return trans.NewLarkTransfer(config.Name, config.URL, config.Prefix)
	case trans.TypeFile:
		return trans.NewFileTransfer(config.Name, config.Dir)
	case trans.TypeConsole:
		return &trans.ConsoleTransfer{
			ID: config.Name,
		}
	default:
		return &trans.NullTransfer{
			ID: config.Name,
		}
	}
}
