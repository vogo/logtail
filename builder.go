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
	"github.com/vogo/logtail/transfer"
)

func buildRouter(s *Server, config *RouterConfig) *Router {
	return NewRouter(s, config.Name, buildMatchers(config.Matchers), buildTransfers(s.runner, config.Transfers))
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

func buildTransfers(runner *Runner, ids []string) []transfer.Transfer {
	transfers := make([]transfer.Transfer, 0, len(ids))

	for _, id := range ids {
		existTransfer, ok := runner.Transfers[id]
		if !ok {
			logger.Errorf("transfer not exists: %s", id)

			continue
		}

		transfers = append(transfers, existTransfer)
	}

	return transfers
}

// nolint:ireturn // return diff transfer implementation.
func buildTransfer(config *TransferConfig) transfer.Transfer {
	switch config.Type {
	case transfer.TypeWebhook:
		return transfer.NewWebhookTransfer(config.Name, config.URL)
	case transfer.TypeDing:
		return transfer.NewDingTransfer(config.Name, config.URL)
	case transfer.TypeLark:
		return transfer.NewLarkTransfer(config.Name, config.URL)
	case transfer.TypeFile:
		return transfer.NewFileTransfer(config.Name, config.Dir)
	case transfer.TypeConsole:
		return &transfer.ConsoleTransfer{
			ID: config.Name,
		}
	default:
		return &transfer.NullTransfer{
			ID: config.Name,
		}
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
