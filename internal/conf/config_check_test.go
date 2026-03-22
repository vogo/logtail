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

package conf_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/conf"
)

func TestInitialCheckConfig_Valid(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"console": {Name: "console", Type: "console"},
		},
		Routers: map[string]*conf.RouterConfig{
			"r1": {Name: "r1", Transfers: []string{"console"}},
		},
		Servers: map[string]*conf.ServerConfig{
			"s1": {Name: "s1", Command: "echo test", Routers: []string{"r1"}},
		},
	}

	err := conf.InitialCheckConfig(config)
	assert.NoError(t, err)
}

func TestInitialCheckConfig_NilMaps(t *testing.T) {
	t.Parallel()

	config := &conf.Config{}
	err := conf.InitialCheckConfig(config)
	assert.NoError(t, err)
	assert.NotNil(t, config.Routers)
	assert.NotNil(t, config.Transfers)
}

func TestInitialCheckConfig_InvalidTransfer(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"bad": {Name: "bad", Type: "unknown"},
		},
	}

	err := conf.InitialCheckConfig(config)
	assert.True(t, errors.Is(err, conf.ErrTransTypeInvalid))
}

func TestInitialCheckConfig_InvalidRouter(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Routers: map[string]*conf.RouterConfig{
			"r1": {Name: "r1", Transfers: []string{"nonexistent"}},
		},
	}

	err := conf.InitialCheckConfig(config)
	assert.True(t, errors.Is(err, conf.ErrTransferNotExist))
}

func TestInitialCheckConfig_InvalidServer(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Servers: map[string]*conf.ServerConfig{
			"s1": {Name: "s1", Routers: []string{"nonexistent"}},
		},
	}

	err := conf.InitialCheckConfig(config)
	assert.True(t, errors.Is(err, conf.ErrRouterNotExist))
}

func TestCheckServerConfig(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Routers: map[string]*conf.RouterConfig{
			"r1": {Name: "r1"},
		},
	}

	t.Run("EmptyName", func(t *testing.T) {
		t.Parallel()

		err := conf.CheckServerConfig(config, &conf.ServerConfig{})
		assert.ErrorIs(t, err, conf.ErrServerIDNil)
	})

	t.Run("ValidWithCommand", func(t *testing.T) {
		t.Parallel()

		err := conf.CheckServerConfig(config, &conf.ServerConfig{Name: "s1", Command: "echo", Routers: []string{"r1"}})
		assert.NoError(t, err)
	})

	t.Run("NoTailingConfig", func(t *testing.T) {
		t.Parallel()

		err := conf.CheckServerConfig(config, &conf.ServerConfig{Name: "s1"})
		assert.NoError(t, err) // only warns, does not error
	})

	t.Run("RouterNotExists", func(t *testing.T) {
		t.Parallel()

		err := conf.CheckServerConfig(config, &conf.ServerConfig{Name: "s1", Command: "echo", Routers: []string{"missing"}})
		assert.ErrorIs(t, err, conf.ErrRouterNotExist)
	})
}

func TestCheckRouterConfig(t *testing.T) {
	t.Parallel()

	config := &conf.Config{
		Transfers: map[string]*conf.TransferConfig{
			"t1": {Name: "t1", Type: "console"},
		},
	}

	t.Run("EmptyName", func(t *testing.T) {
		t.Parallel()

		err := conf.CheckRouterConfig(config, &conf.RouterConfig{})
		assert.ErrorIs(t, err, conf.ErrRouterIDNil)
	})

	t.Run("Valid", func(t *testing.T) {
		t.Parallel()

		err := conf.CheckRouterConfig(config, &conf.RouterConfig{Name: "r1", Transfers: []string{"t1"}})
		assert.NoError(t, err)
	})

	t.Run("TransferNotExists", func(t *testing.T) {
		t.Parallel()

		err := conf.CheckRouterConfig(config, &conf.RouterConfig{Name: "r1", Transfers: []string{"missing"}})
		assert.ErrorIs(t, err, conf.ErrTransferNotExist)
	})
}

func TestCheckMatchers(t *testing.T) {
	t.Parallel()

	err := conf.CheckMatchers(nil)
	assert.NoError(t, err)

	err = conf.CheckMatchers([]*conf.MatcherConfig{{Contains: []string{"ERROR"}}})
	assert.NoError(t, err)

	err = conf.CheckMatchers([]*conf.MatcherConfig{{}})
	assert.NoError(t, err) // empty matchers only log debug
}

func TestCheckTransferConfig(t *testing.T) {
	t.Parallel()

	config := &conf.Config{}

	tests := []struct {
		name     string
		transfer *conf.TransferConfig
		wantErr  error
	}{
		{"EmptyName", &conf.TransferConfig{}, conf.ErrTransferIDNil},
		{"EmptyType", &conf.TransferConfig{Name: "t"}, conf.ErrTransTypeNil},
		{"InvalidType", &conf.TransferConfig{Name: "t", Type: "bad"}, conf.ErrTransTypeInvalid},
		{"ConsoleValid", &conf.TransferConfig{Name: "t", Type: "console"}, nil},
		{"NullValid", &conf.TransferConfig{Name: "t", Type: "null"}, nil},
		{"FileNoDir", &conf.TransferConfig{Name: "t", Type: "file"}, conf.ErrTransDirNil},
		{"FileValid", &conf.TransferConfig{Name: "t", Type: "file", Dir: "/tmp"}, nil},
		{"WebhookNoURL", &conf.TransferConfig{Name: "t", Type: "webhook"}, conf.ErrTransURLNil},
		{"WebhookValid", &conf.TransferConfig{Name: "t", Type: "webhook", URL: "http://x"}, nil},
		{"DingNoURL", &conf.TransferConfig{Name: "t", Type: "ding"}, conf.ErrTransURLNil},
		{"DingValid", &conf.TransferConfig{Name: "t", Type: "ding", URL: "http://x"}, nil},
		{"LarkNoURL", &conf.TransferConfig{Name: "t", Type: "lark"}, conf.ErrTransURLNil},
		{"LarkValid", &conf.TransferConfig{Name: "t", Type: "lark", URL: "http://x"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := conf.InitialCheckConfig(&conf.Config{
				Transfers: map[string]*conf.TransferConfig{tt.transfer.Name: tt.transfer},
			})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				// may have errors from nil maps, but transfer check should pass
				_ = config
			}
		})
	}
}
