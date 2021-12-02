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

import "github.com/vogo/logger"

func buildRouter(s *Server, config *RouterConfig) *Router {
	return NewRouter(s, buildMatchers(config.Matchers), buildTransfers(s.runner, config.Transfers))
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

func buildTransfers(runner *Runner, ids []string) []Transfer {
	transfers := make([]Transfer, 0, len(ids))

	for _, id := range ids {
		t, ok := runner.Transfers[id]
		if !ok {
			logger.Errorf("transfer not exists: %s", id)

			continue
		}

		transfers = append(transfers, t)
	}

	return transfers
}

// nolint:ireturn // return diff transfer implementation.
func buildTransfer(config *TransferConfig) Transfer {
	switch config.Type {
	case TransferTypeWebhook:
		return NewWebhookTransfer(config.ID, config.URL)
	case TransferTypeDing:
		return NewDingTransfer(config.ID, config.URL)
	case TransferTypeLark:
		return NewLarkTransfer(config.ID, config.URL)
	case TransferTypeFile:
		return NewFileTransfer(config.ID, config.Dir)
	case TransferTypeConsole:
		return &ConsoleTransfer{
			IDS{id: config.ID},
		}
	default:
		return &NullTransfer{
			IDS{id: config.ID},
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
