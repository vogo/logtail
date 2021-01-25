package logtail

import (
	"github.com/vogo/logger"
)

// StopLogtail start config servers.
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

	NewServer(c, config)
}

func buildRouter(config *RouterConfig) *Router {
	return newRouter(config.ID, buildMatchers(config.Matchers), buildTransfers(config.Transfers))
}

func buildMatchers(matcherConfigs []*MatcherConfig) []Matcher {
	var matchers []Matcher

	for _, matchConfig := range matcherConfigs {
		matcher := buildMatcher(matchConfig)
		if matcher != nil {
			matchers = append(matchers, matcher)
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

func buildTransfer(config *TransferConfig) Transfer {
	if config.Type == TransferTypeWebhook {
		NewWebhookTransfer(config.URL)
	}

	if config.Type == TransferTypeDing {
		return NewDingTransfer(config.URL)
	}

	return &ConsoleTransfer{}
}

func buildMatcher(config *MatcherConfig) *ContainsMatcher {
	if config.MatchContains == "" {
		return nil
	}

	return NewContainsMatcher(config.MatchContains)
}
