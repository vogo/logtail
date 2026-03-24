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

package cmd_tail_file_transfer_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vogo/logtail/integrations/helper"
)

func TestCmdTailFileTransfer(t *testing.T) {
	binary := helper.BuildBinary(t)
	dir := helper.TempDir(t, "cmd-tail-file")

	outputDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}

	config := map[string]any{
		"loglevel": "ERROR",
		"transfers": map[string]any{
			"file-out": map[string]any{
				"type": "file",
				"dir":  outputDir,
			},
		},
		"routers": map[string]any{
			"all": map[string]any{
				"transfers": []string{"file-out"},
			},
		},
		"servers": map[string]any{
			"s1": map[string]any{
				"command": "printf 'line1\\nline2\\nline3\\n'",
				"routers": []string{"all"},
			},
		},
	}

	configPath := helper.WriteConfig(t, dir, config)
	proc := helper.RunLogtail(t, binary, "-file", configPath)

	filePath := helper.WaitForFileWithPrefix(t, outputDir, "file-out-", 10*time.Second)
	proc.Stop()
	helper.AssertFileContains(t, filePath, "line1")
	helper.AssertFileContains(t, filePath, "line2")
	helper.AssertFileContains(t, filePath, "line3")
}
