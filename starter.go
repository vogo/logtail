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
	"github.com/vogo/logger"
)

// StartLogtail start config servers.
func StartLogtail(config *Config) {
	defaultFormat = config.DefaultFormat

	for _, serverConfig := range config.Servers {
		startServer(config, serverConfig)
	}
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

func buildRouter(s *Server, config *RouterConfig) *Router {
	return NewRouter(s, buildMatchers(config.Matchers), buildTransfers(config.Transfers))
}

func buildMatchers(matcherConfigs []*MatcherConfig) []Matcher {
	var matchers []Matcher

	for _, matchConfig := range matcherConfigs {
		m := buildMatcher(matchConfig)
		if len(m) > 0 {
			matchers = append(matchers, m...)
		}
	}

	return matchers
}

func buildTransfers(transferConfigs []*TransferConfig) []Transfer {
	transfers := make([]Transfer, len(transferConfigs))

	for i, transferConfig := range transferConfigs {
		transfers[i] = buildTransfer(transferConfig)
	}

	return transfers
}

// nolint:ireturn // return diff transfer implementation.
func buildTransfer(config *TransferConfig) Transfer {
	switch config.Type {
	case TransferTypeWebhook:
		return NewWebhookTransfer(config.URL)
	case TransferTypeDing:
		return NewDingTransfer(config.URL)
	case TransferTypeLark:
		return NewLarkTransfer(config.URL)
	case TransferTypeFile:
		return NewFileTransfer(config.Dir)
	case TransferTypeConsole:
		return &ConsoleTransfer{}
	default:
		return &NullTransfer{}
	}
}

func buildMatcher(config *MatcherConfig) []Matcher {
	matchers := make([]Matcher, len(config.Contains)+len(config.NotContains))

	for i, contains := range config.Contains {
		matchers[i] = NewContainsMatcher(contains, true)
	}

	containsLen := len(config.Contains)

	for i, contains := range config.NotContains {
		matchers[i+containsLen] = NewContainsMatcher(contains, false)
	}

	return matchers
}
