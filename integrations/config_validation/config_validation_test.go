//go:build integration

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

package config_validation_test

import (
	"testing"
	"time"

	"github.com/vogo/logtail/integrations/helper"
)

func TestConfigValidation(t *testing.T) {
	binary := helper.BuildBinary(t)

	t.Run("MissingRouter", func(t *testing.T) {
		dir := helper.TempDir(t, "cfg-missing-router")

		config := map[string]any{
			"transfers": map[string]any{
				"console": map[string]any{
					"type": "console",
				},
			},
			"routers": map[string]any{
				"all": map[string]any{
					"transfers": []string{"console"},
				},
			},
			"servers": map[string]any{
				"s1": map[string]any{
					"command": "printf 'test\n'",
					"routers": []string{"nonexistent"},
				},
			},
		}

		configPath := helper.WriteConfig(t, dir, config)
		proc := helper.RunLogtail(t, binary, "-file", configPath)

		if err := proc.Wait(5 * time.Second); err != nil {
			t.Fatalf("process did not exit: %v", err)
		}

		if proc.ExitCode() == 0 {
			t.Error("expected non-zero exit code")
		}

		helper.AssertStderrContains(t, proc, "router not exists")
	})

	t.Run("MissingTransfer", func(t *testing.T) {
		dir := helper.TempDir(t, "cfg-missing-transfer")

		config := map[string]any{
			"transfers": map[string]any{
				"console": map[string]any{
					"type": "console",
				},
			},
			"routers": map[string]any{
				"all": map[string]any{
					"transfers": []string{"nonexistent"},
				},
			},
			"servers": map[string]any{
				"s1": map[string]any{
					"command": "printf 'test\n'",
					"routers": []string{"all"},
				},
			},
		}

		configPath := helper.WriteConfig(t, dir, config)
		proc := helper.RunLogtail(t, binary, "-file", configPath)

		if err := proc.Wait(5 * time.Second); err != nil {
			t.Fatalf("process did not exit: %v", err)
		}

		if proc.ExitCode() == 0 {
			t.Error("expected non-zero exit code")
		}

		helper.AssertStderrContains(t, proc, "transfer not exists")
	})

	t.Run("InvalidTransferType", func(t *testing.T) {
		dir := helper.TempDir(t, "cfg-invalid-type")

		config := map[string]any{
			"transfers": map[string]any{
				"bad": map[string]any{
					"type": "unknown",
				},
			},
			"routers": map[string]any{
				"all": map[string]any{
					"transfers": []string{"bad"},
				},
			},
			"servers": map[string]any{
				"s1": map[string]any{
					"command": "printf 'test\n'",
					"routers": []string{"all"},
				},
			},
		}

		configPath := helper.WriteConfig(t, dir, config)
		proc := helper.RunLogtail(t, binary, "-file", configPath)

		if err := proc.Wait(5 * time.Second); err != nil {
			t.Fatalf("process did not exit: %v", err)
		}

		if proc.ExitCode() == 0 {
			t.Error("expected non-zero exit code")
		}

		helper.AssertStderrContains(t, proc, "invalid transfer type")
	})

	t.Run("FileTransferWithoutDir", func(t *testing.T) {
		dir := helper.TempDir(t, "cfg-no-dir")

		config := map[string]any{
			"transfers": map[string]any{
				"fileout": map[string]any{
					"type": "file",
				},
			},
			"routers": map[string]any{
				"all": map[string]any{
					"transfers": []string{"fileout"},
				},
			},
			"servers": map[string]any{
				"s1": map[string]any{
					"command": "printf 'test\n'",
					"routers": []string{"all"},
				},
			},
		}

		configPath := helper.WriteConfig(t, dir, config)
		proc := helper.RunLogtail(t, binary, "-file", configPath)

		if err := proc.Wait(5 * time.Second); err != nil {
			t.Fatalf("process did not exit: %v", err)
		}

		if proc.ExitCode() == 0 {
			t.Error("expected non-zero exit code")
		}

		helper.AssertStderrContains(t, proc, "transfer dir is nil")
	})
}
