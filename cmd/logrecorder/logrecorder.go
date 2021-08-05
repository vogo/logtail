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

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type recorder struct{}

func (r *recorder) Write(p []byte) (int, error) {
	_, _ = fmt.Fprintf(os.Stdout, "\n----------------\n")

	return os.Stdout.Write(p)
}

func main() {
	c := flag.String("c", "", "command")

	flag.Parse()

	if *c == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cmd := exec.Command("/bin/sh", "-c", *c)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmd.Stdout = &recorder{}
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "command error: %v", err)
	}
}
