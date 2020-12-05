package logtail

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/vogo/vogo/vos"
)

func Start() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			flag.PrintDefaults()
		}
	}()

	flag.Parse()
	config, err := parseConfig()
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}
	vos.LoadUserEnv()
	startLogtail(config)
}

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

	server := &Server{
		id:          config.ID,
		lock:        sync.Mutex{},
		command:     config.Command,
		routerCount: 0,
	}

	serverDB[config.ID] = server
	for _, routerConfig := range config.Routers {
		server.AddRouter(buildRouter(routerConfig))
	}
	// TODO allow to stop command
	startTailCommand(server.command)
}

func buildRouter(config *RouterConfig) *Router {
	var filters []*Filter
	var transfers []Transfer
	for _, filterConfig := range config.Filters {
		filters = append(filters, buildFilter(filterConfig))
	}
	for _, transferConfig := range config.Transfers {
		transfers = append(transfers, buildTransfer(transferConfig))
	}
	var router = NewRouter(filters, transfers)
	return router
}

func buildTransfer(config *TransferConfig) Transfer {
	if config.DingURL != "" {
		return NewWebhookTransfer(config.DingURL)
	}
	return NewWebhookTransfer(config.WebhookURL)
}

func buildFilter(config *FilterConfig) *Filter {
	matcher := NewContainsMatcher(config.MatchContains)
	return &Filter{
		Matcher: matcher,
	}
}
