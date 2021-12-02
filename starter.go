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

// nolint:gochecknoglobals // ignore this
var defaultRunner *Runner

// StartLogtail Start config servers.
func StartLogtail(config *Config) error {
	runner, err := NewRunner(config)
	if err != nil {
		return err
	}

	return StartRunner(runner)
}

// StartRunner Start config servers.
func StartRunner(runner *Runner) error {
	if defaultRunner != nil {
		defaultRunner.Stop()
	}

	defaultRunner = runner

	return defaultRunner.Start()
}

// StopLogtail stop logtail.
func StopLogtail() error {
	if defaultRunner != nil {
		defaultRunner.Stop()
		defaultRunner = nil
	}

	return nil
}
