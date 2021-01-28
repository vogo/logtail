package main

import (
	"github.com/vogo/logtail"
	"github.com/vogo/logtail/repeater"
)

func main() {
	config := &logtail.Config{
		DefaultFormat: &logtail.Format{Prefix: "!!!!-!!-!!"},
		DefaultRouters: []*logtail.RouterConfig{
			{
				ID: "test-router",
				Matchers: []*logtail.MatcherConfig{
					{
						MatchContains: "ERROR",
					},
				},
				Transfers: []*logtail.TransferConfig{
					{
						Type: "console",
					},
					{
						Type: "ding",
						URL:  "http://localhost:55321",
					},
				},
			},
		},
		Servers: []*logtail.ServerConfig{
			{
				ID: "server-1",
			},
		},
	}

	server := logtail.NewServer(config, config.Servers[0])

	c := make(chan []byte)
	go repeater.Repeat("/tmp/logtail/logs/test.log", c)

	for {
		b := <-c
		_ = server.Fire(b)
	}
}
