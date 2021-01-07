package logtail

import (
	"fmt"
	"net/http"

	"github.com/vogo/logger"
)

func startLogtail(config *Config) {
	defaultFormat = config.DefaultFormat

	restartRouters(&defaultRouters, config.DefaultRouters)
	restartRouters(&globalRouters, config.GlobalRouters)

	for _, serverConfig := range config.Servers {
		startServer(serverConfig)
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), &httpHandler{}); err != nil {
		panic(err)
	}
}

func stopServers() {
	for _, s := range serverDB {
		if err := s.Stop(); err != nil {
			logger.Errorf("server %s stop error: %+v", s.id, err)
		}
	}
}

func restartRouters(routers *[]*Router, routerConfigs []*RouterConfig) {
	if len(*routers) > 0 {
		for _, r := range *routers {
			r.Stop()
		}

		*routers = nil
	}

	if len(routerConfigs) > 0 {
		for _, routerConfig := range routerConfigs {
			r := buildRouter(routerConfig)
			*routers = append(*routers, r)

			go func() {
				r.Start()
			}()
		}
	}
}

func startServer(config *ServerConfig) {
	serverDBLock.Lock()
	defer serverDBLock.Unlock()

	server := NewServer(config)
	server.Start()

	for _, routerConfig := range config.Routers {
		server.StartRouter(buildRouter(routerConfig))
	}
}

func buildRouter(config *RouterConfig) *Router {
	return NewRouter(config.ID, buildMatchers(config.Matchers), buildTransfers(config.Transfers))
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
