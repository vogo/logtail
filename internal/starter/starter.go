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

package starter

import (
	"fmt"

	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/tail"
	"github.com/vogo/vogo/vos"
)

// StartLogtail creates a tailer from config and starts it.
func StartLogtail(config *conf.Config) (*tail.Tailer, error) {
	tailer, err := tail.NewTailer(config)
	if err != nil {
		return nil, err
	}

	if err := tailer.Start(); err != nil {
		return nil, err
	}

	return tailer, nil
}

// Start parses command config and starts logtail.
func Start() (*tail.Tailer, error) {
	config, parseErr := conf.ParseConfig()
	if parseErr != nil {
		return nil, fmt.Errorf("parse config: %w", parseErr)
	}

	vos.LoadUserEnv()

	return StartLogtail(config)
}
