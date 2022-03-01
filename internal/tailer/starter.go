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

package tailer

import (
	"flag"
	"fmt"
	"os"

	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/vogo/vos"
)

// DefaultRunner the default runner.
// nolint:gochecknoglobals // ignore this
var DefaultRunner *Runner

// StartLogtail Start config servers.
func StartLogtail(config *conf.Config) error {
	runner, err := NewRunner(config)
	if err != nil {
		return err
	}

	return StartRunner(runner)
}

// StartRunner Start config servers.
func StartRunner(runner *Runner) error {
	if DefaultRunner != nil {
		DefaultRunner.Stop()
	}

	DefaultRunner = runner

	return DefaultRunner.Start()
}

// StopLogtail stop logtail.
func StopLogtail() error {
	if DefaultRunner != nil {
		DefaultRunner.Stop()
		DefaultRunner = nil
	}

	return nil
}

// Start parse command config, and start logtail servers with http listener.
func Start() *Runner {
	config, parseErr := conf.ParseConfig()
	if parseErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, parseErr)

		flag.PrintDefaults()
		os.Exit(1)
	}

	vos.LoadUserEnv()

	runner, err := NewRunner(config)
	if err != nil {
		panic(err)
	}

	go func() {
		if startErr := StartRunner(runner); startErr != nil {
			panic(startErr)
		}
	}()

	return runner
}
