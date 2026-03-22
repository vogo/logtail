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

package cli_flags_test

import (
	"testing"
	"time"

	"github.com/vogo/logtail/integrations/helper"
)

func TestCLIFlags(t *testing.T) {
	binary := helper.BuildBinary(t)

	t.Run("CmdAloneFails", func(t *testing.T) {
		proc := helper.RunLogtail(t, binary, "-cmd", "printf 'hello\n'")

		if err := proc.Wait(5 * time.Second); err != nil {
			t.Fatalf("process did not exit: %v", err)
		}

		if proc.ExitCode() == 0 {
			t.Error("expected non-zero exit code")
		}

		helper.AssertStderrContains(t, proc, "router not exists")
	})

	t.Run("CmdWithMatchContainsNoURLFails", func(t *testing.T) {
		proc := helper.RunLogtail(t, binary, "-cmd", "printf 'ERROR critical\n'", "-match-contains", "ERROR")

		if err := proc.Wait(5 * time.Second); err != nil {
			t.Fatalf("process did not exit: %v", err)
		}

		if proc.ExitCode() == 0 {
			t.Error("expected non-zero exit code")
		}

		helper.AssertStderrContains(t, proc, "router id is nil")
	})
}
