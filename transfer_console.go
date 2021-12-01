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

package logtail

import (
	"os"
)

const TransferTypeConsole = "console"

type ConsoleTransfer struct {
	IDS
}

func (d *ConsoleTransfer) Trans(serverID string, data ...[]byte) error {
	for _, b := range data {
		_, _ = os.Stdout.Write(b)

		n := len(b)
		if n > 0 && b[n-1] != '\n' {
			_, _ = os.Stdout.Write([]byte{'\n'})
		}
	}

	return nil
}

func (d *ConsoleTransfer) start() error { return nil }

func (d *ConsoleTransfer) stop() error { return nil }

func (d *ConsoleTransfer) Visit(t Transfer) {
}
