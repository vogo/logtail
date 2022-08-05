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

	for i := 0; i < 100; i++ {
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

	for i := 0; i < 1000; i++ {
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

	defer os.RemoveAll(workDir)

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

	assert.Nil(t, starter.StartLogtail(config))

	<-time.After(time.Second * 2)

	_ = starter.StopLogtail()

	<-time.After(time.Second * 2)

	config.Servers[testServerName].CommandGen = commandGen

	assert.Nil(t, starter.StartLogtail(config))

	<-time.After(time.Second * 2)

	_ = starter.StopLogtail()
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
