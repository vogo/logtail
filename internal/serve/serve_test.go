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

package serve_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/serve"
	"github.com/vogo/logtail/internal/trans"
)

func TestNewRawServer(t *testing.T) {
	t.Parallel()

	server := serve.NewRawServer("test-server")
	assert.Equal(t, "test-server", server.ID)
	assert.NotNil(t, server.Workers)
	assert.Empty(t, server.Workers)
}

func TestServerStartStop_Command(t *testing.T) {
	t.Parallel()

	server := serve.NewRawServer("cmd-server")
	server.RouterConfigsFunc = func() []*conf.RouterConfig { return nil }
	server.TransferMatcher = func(_ []string) []trans.Transfer { return nil }

	serverConfig := &conf.ServerConfig{
		Name:    "cmd-server",
		Command: "echo hello",
	}

	server.Start(serverConfig)

	time.Sleep(100 * time.Millisecond)

	err := server.Stop()
	assert.NoError(t, err)
}

func TestServerStartStop_MultipleCommands(t *testing.T) {
	t.Parallel()

	server := serve.NewRawServer("multi-cmd")
	server.RouterConfigsFunc = func() []*conf.RouterConfig { return nil }
	server.TransferMatcher = func(_ []string) []trans.Transfer { return nil }

	serverConfig := &conf.ServerConfig{
		Name:     "multi-cmd",
		Commands: "echo line1\necho line2",
	}

	server.Start(serverConfig)

	time.Sleep(100 * time.Millisecond)

	err := server.Stop()
	assert.NoError(t, err)
}

func TestServerStartStop_NoCommand(t *testing.T) {
	t.Parallel()

	server := serve.NewRawServer("no-cmd")
	server.RouterConfigsFunc = func() []*conf.RouterConfig { return nil }
	server.TransferMatcher = func(_ []string) []trans.Transfer { return nil }

	serverConfig := &conf.ServerConfig{
		Name: "no-cmd",
	}

	server.Start(serverConfig)

	time.Sleep(50 * time.Millisecond)

	err := server.Stop()
	assert.NoError(t, err)
}

func TestServerStopWorkers(t *testing.T) {
	t.Parallel()

	server := serve.NewRawServer("sw")
	server.RouterConfigsFunc = func() []*conf.RouterConfig { return nil }
	server.TransferMatcher = func(_ []string) []trans.Transfer { return nil }

	serverConfig := &conf.ServerConfig{
		Name:    "sw",
		Command: "echo test",
	}

	server.Start(serverConfig)

	time.Sleep(100 * time.Millisecond)

	require.NoError(t, server.Stop())
	assert.Empty(t, server.Workers)
}

func TestServerStartFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	server := serve.NewRawServer("file-server")
	server.RouterConfigsFunc = func() []*conf.RouterConfig { return nil }
	server.TransferMatcher = func(_ []string) []trans.Transfer { return nil }

	serverConfig := &conf.ServerConfig{
		Name: "file-server",
		File: &conf.FileConfig{
			Path: dir + "/nonexistent.log",
		},
	}

	server.Start(serverConfig)

	time.Sleep(100 * time.Millisecond)

	require.NoError(t, server.Stop())
}
