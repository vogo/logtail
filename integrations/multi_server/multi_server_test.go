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

package multi_server_test

import (
	"testing"
	"time"

	"github.com/vogo/logtail/integrations/helper"
)

func TestMultiServer(t *testing.T) {
	binary := helper.BuildBinary(t)
	dir := helper.TempDir(t, "multi-server")

	config := map[string]any{
		"loglevel": "ERROR",
		"transfers": map[string]any{
			"console": map[string]any{
				"type": "console",
			},
		},
		"routers": map[string]any{
			"r-error": map[string]any{
				"transfers": []string{"console"},
				"matchers": []map[string]any{
					{
						"contains": []string{"ERROR"},
					},
				},
			},
			"r-warn": map[string]any{
				"transfers": []string{"console"},
				"matchers": []map[string]any{
					{
						"contains": []string{"WARN"},
					},
				},
			},
		},
		"servers": map[string]any{
			"server1": map[string]any{
				"command": "printf 'ERROR server1-error\\n'",
				"routers": []string{"r-error"},
			},
			"server2": map[string]any{
				"command": "printf 'WARN server2-warning\\n'",
				"routers": []string{"r-warn"},
			},
		},
	}

	configPath := helper.WriteConfig(t, dir, config)
	proc := helper.RunLogtail(t, binary, "-file", configPath)

	time.Sleep(2 * time.Second)
	proc.Stop()

	helper.AssertStdoutContains(t, proc, "server1-error")
	helper.AssertStdoutContains(t, proc, "server2-warning")
}
