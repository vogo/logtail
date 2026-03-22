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

package file_watch_console_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vogo/logtail/integrations/helper"
)

func TestFileWatchConsole(t *testing.T) {
	binary := helper.BuildBinary(t)
	dir := helper.TempDir(t, "file-watch-console")

	logPath := filepath.Join(dir, "app.log")
	if err := os.WriteFile(logPath, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	config := map[string]any{
		"loglevel": "ERROR",
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
				"file": map[string]any{
					"path": logPath,
				},
				"routers": []string{"all"},
			},
		},
	}

	configPath := helper.WriteConfig(t, dir, config)
	proc := helper.RunLogtail(t, binary, "-file", configPath)

	time.Sleep(1 * time.Second)
	helper.AppendToFile(t, logPath, "new-appended-line\n")
	time.Sleep(2 * time.Second)

	proc.Stop()

	helper.AssertStdoutContains(t, proc, "new-appended-line")
}
