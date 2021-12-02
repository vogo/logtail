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

// check the config and fill some default values.
func initialCheckConfig(config *Config) error {
	config.routerMap = make(map[string]*RouterConfig, defaultMapSize)
	config.transferMap = make(map[string]*TransferConfig, defaultMapSize)

	if config.Port == 0 {
		config.Port = DefaultServerPort
	}

	if len(config.Servers) == 0 {
		return ErrNoServerConfig
	}

	if len(config.Transfers) == 0 {
		return ErrTransferNotExist
	}

	for _, t := range config.Transfers {
		if transferErr := checkTransferConfig(config, t); transferErr != nil {
			return transferErr
		}
	}

	if routerErr := checkRouterConfigs(config, config.Routers); routerErr != nil {
		return routerErr
	}

	if refErr := checkRouterRef(config, config.DefaultRouters); refErr != nil {
		return refErr
	}

	if globalRefErr := checkRouterRef(config, config.GlobalRouters); globalRefErr != nil {
		return globalRefErr
	}

	for _, server := range config.Servers {
		if serverErr := checkServerConfig(config, server); serverErr != nil {
			return serverErr
		}
	}

	return nil
}

func checkServerConfig(config *Config, server *ServerConfig) error {
	if server.ID == "" {
		return ErrServerIDNil
	}

	if server.Command == "" && server.Commands == "" && server.CommandGen == "" && server.File == nil {
		logger.Warnf("%v for server %s", ErrNoTailingConfig, server.ID)
	}

	if err := checkRouterConfigs(config, server.Routers); err != nil {
		return err
	}

	return nil
}

func checkRouterConfigs(config *Config, routers []*RouterConfig) error {
	for _, router := range routers {
		if err := validateRouterConfig(config, router); err != nil {
			return err
		}
	}

	return nil
}

func validateRouterConfig(config *Config, router *RouterConfig) error {
	if router.ID == "" {
		return ErrRouterIDNil
	}

	if _, ok := config.routerMap[router.ID]; ok {
		return fmt.Errorf("%w: %s %s", ErrDuplicatedConfig, "router", router.ID)
	}

	if err := validateMatchers(router.Matchers); err != nil {
		return err
	}

	if err := validateTransferRef(config, router.Transfers); err != nil {
		return err
	}

	config.routerMap[router.ID] = router

	return nil
}

func checkRouterRef(config *Config, routers []string) error {
	for _, r := range routers {
		if _, ok := config.routerMap[r]; !ok {
			return fmt.Errorf("%w: %s", ErrRouterNotExist, r)
		}
	}

	return nil
}

func validateMatchers(matchers []*MatcherConfig) error {
	if len(matchers) > 0 {
		for _, filter := range matchers {
			if err := validateMatchConfig(filter); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateTransferRef(config *Config, transfers []string) error {
	if len(transfers) > 0 {
		for _, id := range transfers {
			if _, ok := config.transferMap[id]; !ok {
				return fmt.Errorf("%w: %s", ErrTransferNotExist, id)
			}
		}
	}

	return nil
}

func checkTransferConfig(config *Config, transfer *TransferConfig) error {
	if transfer.ID == "" {
		return ErrTransferIDNil
	}

	if transfer.Type == "" {
		return ErrTransTypeNil
	}

	switch transfer.Type {
	case TransferTypeWebhook, TransferTypeDing, TransferTypeLark:
		if transfer.URL == "" {
			return ErrTransURLNil
		}
	case TransferTypeFile:
		if transfer.Dir == "" {
			return ErrTransDirNil
		}
	case TransferTypeConsole, TransferTypeNull:
		break
	default:
		return fmt.Errorf("%w: %s", ErrTransTypeInvalid, transfer.Type)
	}

	config.transferMap[transfer.ID] = transfer

	return nil
}

func validateMatchConfig(config *MatcherConfig) error {
	if len(config.Contains) == 0 && len(config.NotContains) == 0 {
		logger.Debugf("match contains is nil")
	}

	return nil
}
