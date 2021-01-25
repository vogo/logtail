package logtail_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	_ = server.Fire([]byte(`2020-11-11 ERROR test1
 follow1
 follow2`))

	_ = server.Fire([]byte(`2020-11-11 ERROR test2 "中文"
 follow3
 follow4`))

	_ = server.Fire([]byte(`2020-11-11 ERROR test3
 follow5
 follow6`))

	_ = server.Fire([]byte(`follow7
 follow8
2020-11-11 ERROR test4`))

	_ = server.Fire([]byte(`follow5
follow9`))

	<-time.After(time.Second)
}

func TestServerCommands(t *testing.T) {
	workDir := filepath.Join(os.TempDir(), "test_logtail_dir")
	assert.NoError(t, os.MkdirAll(workDir, os.ModePerm))

	defer os.RemoveAll(workDir)

	log1 := filepath.Join(workDir, "log1.txt")
	log2 := filepath.Join(workDir, "log2.txt")

	assert.NoError(t, ioutil.WriteFile(log1, []byte(`2020-11-11 ERROR test1
 follow1
 follow2`), 0600))

	assert.NoError(t, ioutil.WriteFile(log2, []byte(`2020-11-11 ERROR test2 "中文"
 follow3
 follow4`), 0600))

	commands := fmt.Sprintf("tail -f %s\ntail -f %s", log1, log2)
	commandGen := fmt.Sprintf("echo \"tail -f %s\ntail -f %s\"", log1, log2)

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
				ID:       "server-1",
				Commands: commands,
			},
		},
	}

	server := logtail.NewServer(config, config.Servers[0])

	<-time.After(time.Second * 2)

	_ = server.Stop()

	config.Servers[0].CommandGen = commandGen

	logtail.NewServer(config, config.Servers[0])

	<-time.After(time.Second * 2)
}
