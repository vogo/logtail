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

package tail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/tail"
)

func validConfig() *conf.Config {
	return &conf.Config{
		LogLevel: "ERROR",
		Transfers: map[string]*conf.TransferConfig{
			"console": {Name: "console", Type: "console"},
			"null":    {Name: "null", Type: "null"},
		},
		Routers: map[string]*conf.RouterConfig{
			"r1": {Name: "r1", Transfers: []string{"console"}},
		},
		Servers: map[string]*conf.ServerConfig{},
	}
}

func TestNewTailer(t *testing.T) {
	t.Parallel()

	tailer, err := tail.NewTailer(validConfig())
	require.NoError(t, err)
	assert.NotNil(t, tailer)
	assert.NotNil(t, tailer.Servers)
	assert.NotNil(t, tailer.Transfers)
}

func TestNewTailer_InvalidConfig(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"bad": {Name: "bad", Type: "unknown"},
		},
	}

	_, err := tail.NewTailer(config)
	assert.Error(t, err)
}

func TestTailerStartTransfers(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	err = tailer.StartTransfers()
	assert.NoError(t, err)
	assert.Len(t, tailer.Transfers, 2)

	tailer.Stop()
}

func TestBuildTransfer_Console(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{Name: "c", Type: "console"})
	assert.Equal(t, "c", transfer.Name())
}

func TestBuildTransfer_Null(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{Name: "n", Type: "null"})
	assert.Equal(t, "n", transfer.Name())
}

func TestBuildTransfer_Unknown(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{Name: "u", Type: "unknown"})
	assert.Equal(t, "u", transfer.Name()) // defaults to NullTransfer
}

func TestBuildTransfer_File(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{Name: "f", Type: "file", Dir: t.TempDir()})
	assert.Equal(t, "f", transfer.Name())
}

func TestTailerIsTransferUsing(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"console": {Name: "console", Type: "console"},
		},
		Routers: map[string]*conf.RouterConfig{
			"r1": {Name: "r1", Transfers: []string{"console"}},
		},
		Servers: map[string]*conf.ServerConfig{},
	}

	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	assert.True(t, tailer.IsTransferUsing("console"))
	assert.False(t, tailer.IsTransferUsing("nonexistent"))
}

func TestTailerRemoveTransfer_InUse(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	require.NoError(t, tailer.StartTransfers())

	err = tailer.RemoveTransfer("console")
	assert.ErrorIs(t, err, conf.ErrTransferUsing)

	tailer.Stop()
}

func TestTailerRemoveTransfer_NotInUse(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	require.NoError(t, tailer.StartTransfers())

	err = tailer.RemoveTransfer("null")
	assert.NoError(t, err)
	assert.NotContains(t, tailer.Transfers, "null")

	tailer.Stop()
}

func TestTailerRemoveTransfer_NonExistent(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	err = tailer.RemoveTransfer("nonexistent")
	assert.NoError(t, err)
}

func TestTailerIsRouterUsing(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"console": {Name: "console", Type: "console"},
		},
		Routers: map[string]*conf.RouterConfig{
			"r1": {Name: "r1", Transfers: []string{"console"}},
		},
		Servers: map[string]*conf.ServerConfig{
			"s1": {Name: "s1", Command: "echo", Routers: []string{"r1"}},
		},
	}

	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	assert.True(t, tailer.IsRouterUsing("r1"))
	assert.False(t, tailer.IsRouterUsing("nonexistent"))
}

func TestTailerDeleteRouter_InUse(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"console": {Name: "console", Type: "console"},
		},
		Routers: map[string]*conf.RouterConfig{
			"r1": {Name: "r1", Transfers: []string{"console"}},
		},
		Servers: map[string]*conf.ServerConfig{
			"s1": {Name: "s1", Command: "echo", Routers: []string{"r1"}},
		},
	}

	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	err = tailer.DeleteRouter("r1")
	assert.ErrorIs(t, err, conf.ErrRouterUsing)
}

func TestTailerDeleteRouter_NotInUse(t *testing.T) {
	t.Parallel()

	config := validConfig()
	delete(config.Servers, "s1") // no servers using r1

	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	err = tailer.DeleteRouter("r1")
	assert.NoError(t, err)
	assert.NotContains(t, tailer.Config.Routers, "r1")
}

func TestTailerDeleteRouter_NonExistent(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	err = tailer.DeleteRouter("nonexistent")
	assert.NoError(t, err)
}

func TestTailerCollectRouterStats_Empty(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	stats := tailer.CollectRouterStats()
	assert.Empty(t, stats)
}

func TestTailerStop_Empty(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	tailer.Stop() // should not panic
}

func TestBuildTransfer_Webhook(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{
		Name: "wh", Type: "webhook", URL: "http://example.com/hook",
	})
	assert.Equal(t, "wh", transfer.Name())
}

func TestBuildTransfer_Ding(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{
		Name: "d", Type: "ding", URL: "http://ding.example.com",
	})
	assert.Equal(t, "d", transfer.Name())
}

func TestBuildTransfer_Lark(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{
		Name: "l", Type: "lark", URL: "http://lark.example.com",
	})
	assert.Equal(t, "l", transfer.Name())
}

func TestBuildTransfer_WithHTTPOptions(t *testing.T) {
	t.Parallel()

	transfer := tail.BuildTransfer(&conf.TransferConfig{
		Name:            "wh",
		Type:            "webhook",
		URL:             "http://example.com",
		MaxIdleConns:    10,
		IdleConnTimeout: "30s",
		RateLimit:       5.0,
		RateBurst:       10,
		BatchSize:       100,
		BatchTimeout:    "5s",
	})
	assert.Equal(t, "wh", transfer.Name())
}

func TestBuildTransfer_WithInvalidDurations(t *testing.T) {
	t.Parallel()

	// Should not panic even with invalid durations
	transfer := tail.BuildTransfer(&conf.TransferConfig{
		Name:            "wh",
		Type:            "webhook",
		URL:             "http://example.com",
		IdleConnTimeout: "invalid",
		BatchTimeout:    "also-invalid",
	})
	assert.Equal(t, "wh", transfer.Name())
}

func TestTailerStart_WithTransfers(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	err = tailer.Start()
	assert.NoError(t, err)
	assert.Len(t, tailer.Transfers, 2)

	tailer.Stop()
}

func TestTailerDeleteServer(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	// Delete non-existent server
	err = tailer.DeleteServer("nonexistent")
	assert.NoError(t, err)
}

func TestTailerAddTransfer(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	require.NoError(t, tailer.StartTransfers())

	err = tailer.AddTransfer(&conf.TransferConfig{Name: "new-null", Type: "null"})
	assert.NoError(t, err)
	assert.Contains(t, tailer.Transfers, "new-null")

	tailer.Stop()
}

func TestTailerStartTransfer_Replace(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	require.NoError(t, tailer.StartTransfers())

	// Replace existing transfer
	_, err = tailer.StartTransfer(&conf.TransferConfig{Name: "console", Type: "console"})
	assert.NoError(t, err)

	tailer.Stop()
}

func TestTailerAddRouter(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	require.NoError(t, tailer.StartTransfers())

	err = tailer.AddRouter(&conf.RouterConfig{Name: "r2", Transfers: []string{"console"}})
	assert.NoError(t, err)
	assert.Contains(t, tailer.Config.Routers, "r2")

	tailer.Stop()
}

func TestTailerAddRouter_InvalidConfig(t *testing.T) {
	t.Parallel()

	config := validConfig()
	tailer, err := tail.NewTailer(config)
	require.NoError(t, err)

	err = tailer.AddRouter(&conf.RouterConfig{}) // empty name
	assert.ErrorIs(t, err, conf.ErrRouterIDNil)
}
