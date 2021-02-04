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

func buildTransfer(config *TransferConfig) Transfer {
	if config.Type == TransferTypeWebhook {
		NewWebhookTransfer(config.URL)
	}

	if config.Type == TransferTypeDing {
		return NewDingTransfer(config.URL)
	}

	if config.Type == TransferTypeFile {
		return NewFileTransfer(config.Dir)
	}

	return &ConsoleTransfer{}
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
