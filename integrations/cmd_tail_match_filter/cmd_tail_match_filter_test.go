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

package cmd_tail_match_filter_test

import (
	"testing"
	"time"

	"github.com/vogo/logtail/integrations/helper"
)

func TestCmdTailMatchFilter(t *testing.T) {
	binary := helper.BuildBinary(t)
	dir := helper.TempDir(t, "cmd-tail-match")

	config := map[string]any{
		"loglevel": "ERROR",
		"transfers": map[string]any{
			"console": map[string]any{
				"type": "console",
			},
		},
		"routers": map[string]any{
			"error-only": map[string]any{
				"transfers": []string{"console"},
				"matchers": []map[string]any{
					{
						"contains":    []string{"ERROR"},
						"notcontains": []string{"IGNORE"},
					},
				},
			},
		},
		"servers": map[string]any{
			"s1": map[string]any{
				"command": "printf '2024-01-01 ERROR real-error\\n2024-01-01 INFO normal-info\\n2024-01-01 ERROR IGNORE skip-this\\n'",
				"routers": []string{"error-only"},
			},
		},
	}

	configPath := helper.WriteConfig(t, dir, config)
	proc := helper.RunLogtail(t, binary, "-file", configPath)

	helper.WaitForStdoutContains(t, proc, "real-error", 10*time.Second)
	proc.Stop()

	helper.AssertStdoutNotContains(t, proc, "normal-info")
	helper.AssertStdoutNotContains(t, proc, "skip-this")
}
