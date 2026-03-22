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

package format_split_test

import (
	"testing"
	"time"

	"github.com/vogo/logtail/integrations/helper"
)

func TestFormatSplit(t *testing.T) {
	binary := helper.BuildBinary(t)
	dir := helper.TempDir(t, "format-split")

	config := map[string]any{
		"loglevel": "ERROR",
		"defaultformat": map[string]any{
			"prefix": "!!!!-!!-!!",
		},
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
						"contains": []string{"ERROR"},
					},
				},
			},
		},
		"servers": map[string]any{
			"s1": map[string]any{
				"command": "printf '2024-01-15 ERROR something failed\\n  stack trace line 1\\n  stack trace line 2\\n2024-01-15 INFO all good\\n'",
				"routers": []string{"error-only"},
			},
		},
	}

	configPath := helper.WriteConfig(t, dir, config)
	proc := helper.RunLogtail(t, binary, "-file", configPath)

	time.Sleep(2 * time.Second)
	proc.Stop()

	helper.AssertStdoutContains(t, proc, "something failed")
	helper.AssertStdoutContains(t, proc, "stack trace line 1")
	helper.AssertStdoutNotContains(t, proc, "all good")
}
