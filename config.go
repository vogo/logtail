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
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"

	"github.com/vogo/fwatch"
	"github.com/vogo/logtail/transfer"
)

const DefaultServerPort = 54321

var (
	ErrNoServerConfig   = errors.New("no server config")
	ErrDuplicatedConfig = errors.New("duplicated config")
	ErrServerIDNil      = errors.New("server id is nil")
	ErrRouterIDNil      = errors.New("router id is nil")
	ErrTransferIDNil    = errors.New("transfer id is nil")
	ErrRouterNotExist   = errors.New("router not exists")
	ErrTransferNotExist = errors.New("transfer not exists")
	ErrTransferUsing    = errors.New("transfer is using")
	ErrRouterUsing      = errors.New("router is using")
	ErrNoTailingConfig  = errors.New("no tailing command/file config")
	ErrTransURLNil      = errors.New("transfer url is nil")
	ErrTransTypeNil     = errors.New("transfer type is nil")
	ErrTransTypeInvalid = errors.New("invalid transfer type")
	ErrTransDirNil      = errors.New("transfer dir is nil")
)

type Config struct {
	Port           int                        `json:"port"`
	LogLevel       string                     `json:"log_level"`
	DefaultFormat  *Format                    `json:"default_format"`
	Transfers      map[string]*TransferConfig `json:"transfers"`
	Routers        map[string]*RouterConfig   `json:"routers"`
	Servers        map[string]*ServerConfig   `json:"servers"`
	DefaultRouters []string                   `json:"default_routers"`
	GlobalRouters  []string                   `json:"global_routers"`
}

func (c *Config) AppendDefaultRouters(configs []*RouterConfig) []*RouterConfig {
	for _, id := range c.DefaultRouters {
		if r, ok := c.Routers[id]; ok {
			configs = append(configs, r)
		}
	}

	return configs
}

func (c *Config) AppendGlobalRouters(configs []*RouterConfig) []*RouterConfig {
	for _, id := range c.GlobalRouters {
		if r, ok := c.Routers[id]; ok {
			configs = append(configs, r)
		}
	}

	return configs
}

func (c *Config) GetRouters(routers []string) []*RouterConfig {
	var configs []*RouterConfig

	for _, id := range routers {
		if r, ok := c.Routers[id]; ok {
			configs = append(configs, r)
		}
	}

	return configs
}

type ServerConfig struct {
	Name    string   `json:"name"`
	Format  *Format  `json:"format"`
	Routers []string `json:"routers"`

	// single command.
	Command string `json:"command"`

	// multiple commands split by new line.
	Commands string `json:"commands"`

	// command to generate multiple commands split by new line.
	CommandGen string `json:"command_gen"`

	// command to generate multiple commands split by new line.
	File *FileConfig `json:"file"`
}

// FileConfig tailing file config.
type FileConfig struct {
	// Path the file or directory to tail.
	Path string `json:"path"`

	// Method watch method,
	// - os: using os file system api to monitor file changes,
	// - timer: interval check file stat to check file changes,
	// For some networks mount devices, can't get file change events for os api,
	// you'd be better to check file stat to know the changes.
	Method fwatch.WatchMethod `json:"method"`

	// only tailing files with the prefix.
	Prefix string `json:"prefix"`

	// only tailing files with the suffix.
	Suffix string `json:"suffix"`

	// Whether include all files in subdirectories recursively.
	Recursive bool `json:"recursive"`
}

type RouterConfig struct {
	Name      string           `json:"name"`
	Matchers  []*MatcherConfig `json:"matchers"`
	Transfers []string         `json:"transfers"`
}

type MatcherConfig struct {
	Contains    []string `json:"contains"`
	NotContains []string `json:"not_contains"`
}

type TransferConfig struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
	Dir  string `json:"dir"`
}

func parseConfig() (cfg *Config, parseErr error) {
	defer func() {
		if err := recover(); err != nil {
			parseErr, _ = err.(error)
		}
	}()

	var (
		file          = flag.String("file", "", "config file")
		port          = flag.Int("port", DefaultServerPort, "tail port")
		command       = flag.String("cmd", "", "tail command")
		matchContains = flag.String("match-contains", "", "a containing string")
		dingURL       = flag.String("ding-url", "", "dingding url")
		webhookURL    = flag.String("webhook-url", "", "webhook url")
	)

	flag.Parse()

	if *file != "" {
		config := &Config{}
		data, fileErr := ioutil.ReadFile(*file)

		if fileErr != nil {
			return nil, fileErr
		}

		if jsonErr := json.Unmarshal(data, config); jsonErr != nil {
			return nil, jsonErr
		}

		return config, nil
	}

	return buildCommandLineConfig(*port, *command, *matchContains, *dingURL, *webhookURL), nil
}

func buildCommandLineConfig(port int, command, matchContains, dingURL, webhookURL string) *Config {
	config := &Config{
		Transfers: make(map[string]*TransferConfig),
		Routers:   make(map[string]*RouterConfig),
		Servers:   make(map[string]*ServerConfig),
	}

	config.Port = port
	serverConfig := &ServerConfig{
		Name:    DefaultID,
		Routers: []string{DefaultID},
	}

	config.Servers[DefaultID] = serverConfig
	serverConfig.Command = command

	if dingURL == "" && webhookURL == "" && matchContains == "" {
		return config
	}

	routerConfig := &RouterConfig{
		Transfers: []string{DefaultID},
	}
	config.Routers[DefaultID] = routerConfig

	if matchContains != "" {
		routerConfig.Matchers = []*MatcherConfig{{
			Contains: []string{matchContains},
		}}
	}

	if dingURL != "" {
		config.Transfers[DefaultID] = &TransferConfig{
			Name: DefaultID,
			Type: transfer.TypeDing,
			URL:  dingURL,
		}
	} else if webhookURL != "" {
		config.Transfers[DefaultID] = &TransferConfig{
			Name: DefaultID,
			Type: transfer.TypeWebhook,
			URL:  webhookURL,
		}
	}

	return config
}
