/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package starter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/match"
	"github.com/vogo/logtail/internal/serve"
	"github.com/vogo/logtail/internal/starter"
	"github.com/vogo/logtail/internal/tail"
	"github.com/vogo/logtail/internal/trans"
)

//nolint:gochecknoglobals // ignore this
var fireData [][]byte

func initFireData() {
	baseText := "long text 数据 long text 数据 long text 数据 long text 数据 long text 数据 long text 数据 long text 数据"

	var longText string

	for range 100 {
		longText += baseText
	}

	fireData = [][]byte{
		[]byte(`2020-11-11 ERROR test1
 follow1
 follow2
`),

		[]byte(`2020-11-11 ERROR test2 "中文"
 follow3
 follow4
`),

		[]byte(`2020-11-11 INFO ` + longText + "\n"),

		[]byte(`2020-11-11 ERROR ` + longText + "\n"),

		[]byte(`2020-11-11 INFO ` + longText + "\n"),

		[]byte(`2020-11-11 ERROR test3
 follow5
 follow6
`),

		[]byte(`follow7
 follow8
2020-11-11 ERROR test4
`),

		[]byte(`follow5
follow9
`),

		[]byte(`2020-11-11 ERROR 6 no TEST should not match` + "\n"),

		[]byte(`2020-11-11 ERROR test7 contains NORMAL so should not match` + "\n"),
	}
}

//nolint:gochecknoglobals // ignore this
var ticker = time.NewTicker(time.Millisecond)

func fireServer(s *serve.Server) {
	for _, b := range fireData {
		<-ticker.C

		_, _ = s.Write(b)
	}
}

func TestServer(t *testing.T) {
	t.Parallel()
	initFireData()

	serverID := "server"
	transType := trans.TypeNull
	config := &conf.Config{
		LogLevel:      "DEBUG",
		DefaultFormat: &match.Format{Prefix: "!!!!-!!-!!"},
		Transfers: map[string]*conf.TransferConfig{
			transType: {
				Name: transType,
				Type: transType,
			},
		},
		Routers: map[string]*conf.RouterConfig{
			"error-router": {
				Name: "error-router",
				Matchers: []*conf.MatcherConfig{
					{
						Contains:    []string{"ERROR", "test"},
						NotContains: []string{"NORMAL"},
					},
				},
				Transfers: []string{transType},
			},
		},
		Servers: map[string]*conf.ServerConfig{
			serverID: {
				Name:    serverID,
				Routers: []string{"error-router"},
			},
		},
	}

	tailer, err := tail.NewTailer(config)
	if err != nil {
		t.Error(err)

		return
	}

	if err = tailer.Start(); err != nil {
		t.Error(err)

		return
	}

	server := tailer.Servers[config.Servers[serverID].Name]

	for range 1000 {
		fireServer(server)
	}

	<-time.After(time.Second)
}

//nolint:gochecknoglobals //ignore this
var testServerName = "svr"

func TestCommands(t *testing.T) {
	t.Parallel()

	workDir := filepath.Join(os.TempDir(), "test_logtail_dir")
	assert.NoError(t, os.MkdirAll(workDir, os.ModePerm))

	defer func() { _ = os.RemoveAll(workDir) }()

	log1 := filepath.Join(workDir, "log1.txt")
	log2 := filepath.Join(workDir, "log2.txt")

	assert.NoError(t, os.WriteFile(log1, []byte(`2020-11-11 ERROR test1
 follow1
 follow2
`), 0o600))

	assert.NoError(t, os.WriteFile(log2, []byte(`2020-11-11 ERROR test2 "中文"
 follow3
 follow4
`), 0o600))

	commands := fmt.Sprintf("tail -f %s\ntail -f %s", log1, log2)
	commandGen := fmt.Sprintf("echo \"tail -f %s\ntail -f %s\"", log1, log2)

	config := testCommandConfig(commands)

	tailer1, err := starter.StartLogtail(config)
	assert.Nil(t, err)
	assert.NotNil(t, tailer1)

	<-time.After(time.Second * 2)

	tailer1.Stop()

	<-time.After(time.Second * 2)

	config.Servers[testServerName].CommandGen = commandGen

	tailer2, err := starter.StartLogtail(config)
	assert.Nil(t, err)
	assert.NotNil(t, tailer2)

	<-time.After(time.Second * 2)

	tailer2.Stop()
}

func TestStartLogtailInvalidConfig(t *testing.T) {
	t.Parallel()

	// Config with server referencing non-existent router should fail validation
	config := &conf.Config{
		Servers: map[string]*conf.ServerConfig{
			"s1": {
				Name:    "s1",
				Command: "echo test",
				Routers: []string{"nonexistent-router"},
			},
		},
	}

	tailer, err := starter.StartLogtail(config)
	assert.Nil(t, tailer)
	assert.NotNil(t, err)
}

func TestStartLogtailValidConfig(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"null": {
				Name: "null",
				Type: trans.TypeNull,
			},
		},
		Routers: map[string]*conf.RouterConfig{
			"r1": {
				Name:      "r1",
				Transfers: []string{"null"},
			},
		},
		Servers: map[string]*conf.ServerConfig{
			"s1": {
				Name:    "s1",
				Command: "echo hello",
				Routers: []string{"r1"},
			},
		},
	}

	tailer, err := starter.StartLogtail(config)
	assert.Nil(t, err)
	assert.NotNil(t, tailer)

	<-time.After(200 * time.Millisecond)

	tailer.Stop()
}

func TestMultipleIndependentTailers(t *testing.T) {
	t.Parallel()

	makeConfig := func(serverName string) *conf.Config {
		return &conf.Config{
			Transfers: map[string]*conf.TransferConfig{
				"null": {
					Name: "null",
					Type: trans.TypeNull,
				},
			},
			Routers: map[string]*conf.RouterConfig{
				"r1": {
					Name:      "r1",
					Transfers: []string{"null"},
				},
			},
			Servers: map[string]*conf.ServerConfig{
				serverName: {
					Name:    serverName,
					Command: "echo hello",
					Routers: []string{"r1"},
				},
			},
		}
	}

	tailer1, err1 := starter.StartLogtail(makeConfig("s1"))
	assert.Nil(t, err1)
	assert.NotNil(t, tailer1)

	tailer2, err2 := starter.StartLogtail(makeConfig("s2"))
	assert.Nil(t, err2)
	assert.NotNil(t, tailer2)

	<-time.After(200 * time.Millisecond)

	// Stop tailer1, tailer2 should remain unaffected
	tailer1.Stop()

	<-time.After(200 * time.Millisecond)

	// tailer2 should still have its server
	assert.NotNil(t, tailer2.Servers["s2"])

	tailer2.Stop()
}

func testCommandConfig(commands string) *conf.Config {
	routerName := "error"

	return &conf.Config{
		DefaultFormat: &match.Format{Prefix: "!!!!-!!-!!"},
		Transfers: map[string]*conf.TransferConfig{
			"console": {
				Name: "console",
				Type: trans.TypeConsole,
			},
		},
		Routers: map[string]*conf.RouterConfig{
			routerName: {
				Name: routerName,
				Matchers: []*conf.MatcherConfig{
					{
						Contains: []string{"ERROR"},
					},
				},
				Transfers: []string{"console"},
			},
		},
		Servers: map[string]*conf.ServerConfig{
			testServerName: {
				Name:     testServerName,
				Commands: commands,
				Routers:  []string{routerName},
			},
		},
	}
}
