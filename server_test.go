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
	t.Parallel()

	config := &logtail.Config{
		DefaultFormat: &logtail.Format{Prefix: "!!!!-!!-!!"},
		DefaultRouters: []*logtail.RouterConfig{
			{
				Matchers: []*logtail.MatcherConfig{
					{
						Contains:    []string{"ERROR", "test"},
						NotContains: []string{"NORMAL"},
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
	server.Start()

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

	_ = server.Fire([]byte(`2020-11-11 ERROR 6 no TEST should not match`))
	_ = server.Fire([]byte(`2020-11-11 ERROR test7 contains NORMAL so should not match`))

	<-time.After(time.Second)
}

func TestServerCommands(t *testing.T) {
	t.Parallel()

	workDir := filepath.Join(os.TempDir(), "test_logtail_dir")
	assert.NoError(t, os.MkdirAll(workDir, os.ModePerm))

	defer os.RemoveAll(workDir)

	log1 := filepath.Join(workDir, "log1.txt")
	log2 := filepath.Join(workDir, "log2.txt")

	assert.NoError(t, ioutil.WriteFile(log1, []byte(`2020-11-11 ERROR test1
 follow1
 follow2`), 0o600))

	assert.NoError(t, ioutil.WriteFile(log2, []byte(`2020-11-11 ERROR test2 "中文"
 follow3
 follow4`), 0o600))

	commands := fmt.Sprintf("tail -f %s\ntail -f %s", log1, log2)
	commandGen := fmt.Sprintf("echo \"tail -f %s\ntail -f %s\"", log1, log2)

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
				},
			},
		},
		Servers: []*logtail.ServerConfig{
			{
				ID:       "server-test",
				Commands: commands,
			},
		},
	}

	logtail.StartLogtail(config)

	<-time.After(time.Second * 2)

	config.Servers[0].CommandGen = commandGen

	logtail.StartLogtail(config)

	<-time.After(time.Second * 2)

	logtail.StopLogtail()
}
