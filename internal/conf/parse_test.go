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

package conf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vogo/logtail/internal/consts"
	"github.com/vogo/logtail/internal/trans"
)

func TestBuildEmptyConfig(t *testing.T) {
	t.Parallel()

	config := buildEmptyConfig()
	assert.Equal(t, "INFO", config.LogLevel)
	assert.NotNil(t, config.Transfers)
	assert.NotNil(t, config.Routers)
	assert.NotNil(t, config.Servers)
	assert.Empty(t, config.Transfers)
	assert.Empty(t, config.Routers)
	assert.Empty(t, config.Servers)
}

func TestBuildCommandLineConfig_NoFlags(t *testing.T) {
	t.Parallel()

	config := buildCommandLineConfig(0, "echo hello", "", "", "")

	assert.Equal(t, 0, config.Port)
	assert.Len(t, config.Servers, 1)
	assert.Equal(t, "echo hello", config.Servers[consts.DefaultID].Command)
	assert.Empty(t, config.Routers)
	assert.Empty(t, config.Transfers)
}

func TestBuildCommandLineConfig_WithPort(t *testing.T) {
	t.Parallel()

	config := buildCommandLineConfig(8080, "echo hello", "", "", "")

	assert.Equal(t, 8080, config.Port)
}

func TestBuildCommandLineConfig_WithDingURL(t *testing.T) {
	t.Parallel()

	config := buildCommandLineConfig(0, "echo hello", "ERROR", "https://ding.example.com/hook", "")

	assert.Len(t, config.Routers, 1)
	assert.Len(t, config.Transfers, 1)

	transfer := config.Transfers[consts.DefaultID]
	assert.Equal(t, trans.TypeDing, transfer.Type)
	assert.Equal(t, "https://ding.example.com/hook", transfer.URL)

	router := config.Routers[consts.DefaultID]
	assert.Len(t, router.Matchers, 1)
	assert.Equal(t, []string{"ERROR"}, router.Matchers[0].Contains)
}

func TestBuildCommandLineConfig_WithWebhookURL(t *testing.T) {
	t.Parallel()

	config := buildCommandLineConfig(0, "echo hello", "", "", "https://webhook.example.com")

	transfer := config.Transfers[consts.DefaultID]
	assert.Equal(t, trans.TypeWebhook, transfer.Type)
	assert.Equal(t, "https://webhook.example.com", transfer.URL)
}

func TestBuildCommandLineConfig_MatchContainsOnly(t *testing.T) {
	t.Parallel()

	config := buildCommandLineConfig(0, "echo hello", "ERROR", "", "")

	router := config.Routers[consts.DefaultID]
	assert.Len(t, router.Matchers, 1)
	assert.Equal(t, []string{"ERROR"}, router.Matchers[0].Contains)

	// No transfer is added (no ding or webhook URL)
	_, hasTransfer := config.Transfers[consts.DefaultID]
	assert.False(t, hasTransfer)
}

func TestParseFileConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")

	configJSON := `{
		"port": 12345,
		"transfers": {
			"console": {"type": "console"}
		},
		"routers": {
			"r1": {"transfers": ["console"]}
		},
		"servers": {
			"s1": {"command": "echo test", "routers": ["r1"]}
		}
	}`

	require.NoError(t, os.WriteFile(configFile, []byte(configJSON), 0o644))

	config, err := parseFileConfig(configFile)
	require.NoError(t, err)

	assert.Equal(t, 12345, config.Port)

	// Verify names are populated from map keys
	assert.Equal(t, "console", config.Transfers["console"].Name)
	assert.Equal(t, "r1", config.Routers["r1"].Name)
	assert.Equal(t, "s1", config.Servers["s1"].Name)
}

func TestParseFileConfig_InvalidFile(t *testing.T) {
	t.Parallel()

	_, err := parseFileConfig("/nonexistent/path/config.json")
	assert.Error(t, err)
}

func TestParseFileConfig_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")

	require.NoError(t, os.WriteFile(configFile, []byte("{invalid"), 0o644))

	_, err := parseFileConfig(configFile)
	assert.Error(t, err)
}

func TestBuildDefaultConfig_WithExistingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")

	configJSON := `{"port": 9999}`
	require.NoError(t, os.WriteFile(configFile, []byte(configJSON), 0o644))

	config := buildDefaultConfig(configFile)
	assert.Equal(t, 9999, config.Port)
}

func TestBuildDefaultConfig_NoFile(t *testing.T) {
	t.Parallel()

	config := buildDefaultConfig("/nonexistent/.logtail.json")
	assert.NotNil(t, config)
	assert.Equal(t, "INFO", config.LogLevel)
}

func TestConfigGetRouters(t *testing.T) {
	t.Parallel()

	r1 := &RouterConfig{Name: "r1"}
	r2 := &RouterConfig{Name: "r2"}

	config := &Config{
		Routers: map[string]*RouterConfig{
			"r1": r1,
			"r2": r2,
		},
	}

	result := config.GetRouters([]string{"r1", "r2"})
	assert.Len(t, result, 2)

	result = config.GetRouters([]string{"r1", "missing"})
	assert.Len(t, result, 1)

	result = config.GetRouters([]string{"missing"})
	assert.Empty(t, result)
}

func TestConfigSaveToFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.json")

	config := &Config{
		file: configFile,
		Port: 8080,
	}

	config.SaveToFile()

	data, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "8080")
}

func TestConfigSaveToFile_EmptyPath(t *testing.T) {
	t.Parallel()

	config := &Config{}
	config.SaveToFile() // should not panic
}
