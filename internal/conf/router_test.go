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

package conf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vogo/logtail/internal/conf"
)

func TestBuildRouterConfigsFunc(t *testing.T) {
	t.Parallel()

	r1 := &conf.RouterConfig{Name: "r1"}
	r2 := &conf.RouterConfig{Name: "r2"}

	config := &conf.Config{
		Routers: map[string]*conf.RouterConfig{
			"r1": r1,
			"r2": r2,
		},
	}

	serverConfig := &conf.ServerConfig{
		Name:    "s1",
		Routers: []string{"r1", "r2"},
	}

	fn := conf.BuildRouterConfigsFunc(config, serverConfig)
	result := fn()
	assert.Len(t, result, 2)

	// Test with missing router
	serverConfig2 := &conf.ServerConfig{
		Name:    "s2",
		Routers: []string{"r1", "missing"},
	}

	fn2 := conf.BuildRouterConfigsFunc(config, serverConfig2)
	result2 := fn2()
	assert.Len(t, result2, 1)
}
