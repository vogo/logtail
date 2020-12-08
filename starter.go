package logtail

import (
	"fmt"
	"net/http"
)

func startLogtail(config *Config) {
	for _, serverConfig := range config.Servers {
		startServer(serverConfig)
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), &httpHandler{}); err != nil {
		panic(err)
	}
}

func startServer(config *ServerConfig) {
	serverDBLock.Lock()
	defer serverDBLock.Unlock()

	server := NewServer(config.ID, config.Command)
	server.Start()

	for _, routerConfig := range config.Routers {
		server.StartRouter(buildRouter(routerConfig))
	}
}

func buildRouter(config *RouterConfig) *Router {
	return NewRouter(buildMatchers(config.Matchers), buildTransfers(config.Transfers))
}

func buildMatchers(matcherConfigs []*MatcherConfig) []Matcher {
	var matchers []Matcher
	for _, matchConfig := range matcherConfigs {
		matchers = append(matchers, buildMatcher(matchConfig))
	}
	return matchers
}

func buildTransfers(transferConfigs []*TransferConfig) []Transfer {
	var transfers []Transfer
	for _, transferConfig := range transferConfigs {
		transfers = append(transfers, buildTransfer(transferConfig))
	}
	return transfers
}

func buildTransfer(config *TransferConfig) Transfer {
	if config.DingURL != "" {
		return NewWebhookTransfer(config.DingURL)
	}
	return NewWebhookTransfer(config.WebhookURL)
}

func buildMatcher(config *MatcherConfig) *ContainsMatcher {
	return NewContainsMatcher(config.MatchContains)
}
