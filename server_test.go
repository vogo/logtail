package logtail_test

import (
	"testing"
	"time"

	"github.com/vogo/logtail"
)

func TestServer(t *testing.T) {
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

	_, _ = server.Write([]byte(`2020-11-11 ERROR test1
 follow1
 follow2`))

	_, _ = server.Write([]byte(`2020-11-11 ERROR test2 "中文"
 follow3
 follow4`))

	_, _ = server.Write([]byte(`2020-11-11 ERROR test3
 follow5
 follow6`))

	_, _ = server.Write([]byte(`follow7
 follow8
2020-11-11 ERROR test4`))

	_, _ = server.Write([]byte(`follow5
follow9`))

	<-time.After(time.Second)
}
