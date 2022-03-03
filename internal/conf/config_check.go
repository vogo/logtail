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

package conf

import (
	"fmt"

	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/logtail/internal/util"
)

// check the config and fill some default values.
func InitialCheckConfig(config *Config) error {
	if config.Routers == nil {
		config.Routers = make(map[string]*RouterConfig, util.DefaultMapSize)
	}

	if config.Transfers == nil {
		config.Transfers = make(map[string]*TransferConfig, util.DefaultMapSize)
	}

	if config.Port == 0 {
		// log for notice
		logger.Infof("port is zero")
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
		if serverErr := CheckServerConfig(config, server); serverErr != nil {
			return serverErr
		}
	}

	return nil
}

func CheckServerConfig(config *Config, server *ServerConfig) error {
	if server.Name == "" {
		return ErrServerIDNil
	}

	if server.Command == "" && server.Commands == "" && server.CommandGen == "" && server.File == nil {
		logger.Warnf("%v for server %s", ErrNoTailingConfig, server.Name)
	}

	return checkRouterRef(config, server.Routers)
}

func checkRouterConfigs(config *Config, routers map[string]*RouterConfig) error {
	for _, router := range routers {
		if err := CheckRouterConfig(config, router); err != nil {
			return err
		}
	}

	return nil
}

func CheckRouterConfig(config *Config, router *RouterConfig) error {
	if router.Name == "" {
		return ErrRouterIDNil
	}

	if err := CheckMatchers(router.Matchers); err != nil {
		return err
	}

	return checkTransferRef(config, router.Transfers)
}

func checkRouterRef(config *Config, routers []string) error {
	for _, r := range routers {
		if _, ok := config.Routers[r]; !ok {
			return fmt.Errorf("%w: %s", ErrRouterNotExist, r)
		}
	}

	return nil
}

func CheckMatchers(matchers []*MatcherConfig) error {
	if len(matchers) > 0 {
		for _, filter := range matchers {
			if err := checkMatchConfig(filter); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkTransferRef(config *Config, ids []string) error {
	if len(ids) > 0 {
		for _, id := range ids {
			if _, ok := config.Transfers[id]; !ok {
				return fmt.Errorf("%w: %s", ErrTransferNotExist, id)
			}
		}
	}

	return nil
}

func checkTransferConfig(_ *Config, transferConfig *TransferConfig) error {
	if transferConfig.Name == "" {
		return ErrTransferIDNil
	}

	if transferConfig.Type == "" {
		return ErrTransTypeNil
	}

	switch transferConfig.Type {
	case trans.TypeWebhook, trans.TypeDing, trans.TypeLark:
		if transferConfig.URL == "" {
			return ErrTransURLNil
		}
	case trans.TypeFile:
		if transferConfig.Dir == "" {
			return ErrTransDirNil
		}
	case trans.TypeConsole, trans.TypeNull:
		break
	default:
		return fmt.Errorf("%w: %s", ErrTransTypeInvalid, transferConfig.Type)
	}

	return nil
}

func checkMatchConfig(config *MatcherConfig) error {
	if len(config.Contains) == 0 && len(config.NotContains) == 0 {
		logger.Debugf("match contains is nil")
	}

	return nil
}
