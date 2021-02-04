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
				Matchers: []*logtail.MatcherConfig{
					{
						Contains: []string{"ERROR"},
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
				ID: "server-test",
			},
		},
	}

	server := logtail.NewServer(config, config.Servers[0])
	server.Start()

	c := make(chan []byte)
	go repeater.Repeat("/Users/gelnyang/temp/logtail/test.log", c)

	for {
		b := <-c
		_ = server.Fire(b)
	}
}
