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
	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/trans"
)

func BuildRouter(s *Server, config *conf.RouterConfig) *Router {
	return NewRouter(s, config.Name, BuildMatchers(config.Matchers), buildTransfers(s.Tailer, config.Transfers))
}

func buildTransfers(runner *Tailer, ids []string) []trans.Transfer {
	transfers := make([]trans.Transfer, 0, len(ids))

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

func NewMatchers(configs []*conf.MatcherConfig) ([]match.Matcher, error) {
	if err := conf.CheckMatchers(configs); err != nil {
		return nil, err
	}

	return BuildMatchers(configs), nil
}

func BuildMatchers(matcherConfigs []*conf.MatcherConfig) []match.Matcher {
	var matchers []match.Matcher

	for _, matchConfig := range matcherConfigs {
		m := buildMatcher(matchConfig)
		if len(m) > 0 {
			matchers = append(matchers, m...)
		}
	}

	return matchers
}

func buildMatcher(config *conf.MatcherConfig) []match.Matcher {
	matchers := make([]match.Matcher, len(config.Contains)+len(config.NotContains))

	for i, contains := range config.Contains {
		matchers[i] = match.NewContainsMatcher(contains, true)
	}

	containsLen := len(config.Contains)

	for i, contains := range config.NotContains {
		matchers[i+containsLen] = match.NewContainsMatcher(contains, false)
	}

	return matchers
}
