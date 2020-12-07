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
	var matchers []Matcher
	var transfers []Transfer
	for _, matchConfig := range config.Matchers {
		matchers = append(matchers, buildMatcher(matchConfig))
	}
	for _, transferConfig := range config.Transfers {
		transfers = append(transfers, buildTransfer(transferConfig))
	}
	var router = NewRouter(matchers, transfers)
	return router
}

func buildTransfer(config *TransferConfig) Transfer {
	if config.DingURL != "" {
		return NewWebhookTransfer(config.DingURL)
	}
	return NewWebhookTransfer(config.WebhookURL)
}

func buildMatcher(config *MatchConfig) *ContainsMatcher {
	return NewContainsMatcher(config.MatchContains)
}
