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
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/consts"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/vogo/vos"
)

//nolint:nonamedreturns //ignore this.
func ParseConfig() (config *Config, parseErr error) {
	defer func() {
		if err := recover(); err != nil {
			parseErr, _ = err.(error)
		}
	}()

	var (
		file          = flag.String("file", "", "config file")
		port          = flag.Int("port", 0, "tail port")
		command       = flag.String("cmd", "", "tail command")
		matchContains = flag.String("match-contains", "", "a containing string")
		dingURL       = flag.String("ding-url", "", "dingding url")
		webhookURL    = flag.String("webhook-url", "", "webhook url")
	)

	flag.Parse()

	if *file != "" {
		return parseFileConfig(*file)
	}

	configFile := filepath.Join(vos.CurrUserHome(), ".logtail.json")

	if *command != "" {
		config = buildCommandLineConfig(*port, *command, *matchContains, *dingURL, *webhookURL)

		config.file = configFile

		return config, nil
	}

	logger.Infof("default config file: %s", configFile)
	config = buildDefaultConfig(configFile)

	if *port > 0 {
		config.Port = *port
	}

	return config, nil
}

func parseFileConfig(f string) (*Config, error) {
	config := &Config{
		file: f,
	}
	data, fileErr := os.ReadFile(f)

	if fileErr != nil {
		return nil, fileErr
	}

	if jsonErr := json.Unmarshal(data, config); jsonErr != nil {
		return nil, jsonErr
	}

	if config.Transfers != nil {
		for k, v := range config.Transfers {
			v.Name = k
		}
	}

	if config.Routers != nil {
		for k, v := range config.Routers {
			v.Name = k
		}
	}

	if config.Servers != nil {
		for k, v := range config.Servers {
			v.Name = k
		}
	}

	return config, nil
}

func buildDefaultConfig(filePath string) *Config {
	if _, err := os.Stat(filePath); err == nil {
		config, fileErr := parseFileConfig(filePath)
		if fileErr == nil {
			return config
		}
	}

	config := buildEmptyConfig()
	config.file = filePath

	return config
}

func buildCommandLineConfig(port int, command, matchContains, dingURL, webhookURL string) *Config {
	config := buildEmptyConfig()

	if port > 0 {
		config.Port = port
	}

	serverConfig := &ServerConfig{
		Name:    consts.DefaultID,
		Routers: []string{consts.DefaultID},
	}

	config.Servers[consts.DefaultID] = serverConfig
	serverConfig.Command = command

	if dingURL == "" && webhookURL == "" && matchContains == "" {
		return config
	}

	routerConfig := &RouterConfig{
		Transfers: []string{consts.DefaultID},
	}
	config.Routers[consts.DefaultID] = routerConfig

	if matchContains != "" {
		routerConfig.Matchers = []*MatcherConfig{{
			Contains: []string{matchContains},
		}}
	}

	if dingURL != "" {
		config.Transfers[consts.DefaultID] = &TransferConfig{
			Name: consts.DefaultID,
			Type: trans.TypeDing,
			URL:  dingURL,
		}
	} else if webhookURL != "" {
		config.Transfers[consts.DefaultID] = &TransferConfig{
			Name: consts.DefaultID,
			Type: trans.TypeWebhook,
			URL:  webhookURL,
		}
	}

	return config
}

func buildEmptyConfig() *Config {
	return &Config{
		LogLevel:  "INFO",
		Transfers: make(map[string]*TransferConfig),
		Routers:   make(map[string]*RouterConfig),
		Servers:   make(map[string]*ServerConfig),
	}
}
