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

// nolint:gochecknoglobals // ignore this
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
 follow2`),

		[]byte(`2020-11-11 ERROR test2 "中文"
 follow3
 follow4`),

		[]byte(`2020-11-11 INFO ` + longText),

		[]byte(`2020-11-11 ERROR ` + longText),

		[]byte(`2020-11-11 INFO ` + longText),

		[]byte(`2020-11-11 ERROR test3
 follow5
 follow6`),

		[]byte(`follow7
 follow8
2020-11-11 ERROR test4`),

		[]byte(`follow5
follow9`),

		[]byte(`2020-11-11 ERROR 6 no TEST should not match`),

		[]byte(`2020-11-11 ERROR test7 contains NORMAL so should not match`),
	}
}

// nolint:gochecknoglobals // ignore this
var ticker = time.NewTicker(time.Millisecond)

func fireServer(s *logtail.Server) {
	for _, b := range fireData {
		<-ticker.C

		_ = s.Fire(b)
	}
}

func TestServer(t *testing.T) {
	t.Parallel()
	initFireData()

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
						Type: "null",
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

	for i := 0; i < 1000; i++ {
		fireServer(server)
	}

	<-time.After(time.Second)
}

func TestCommands(t *testing.T) {
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
