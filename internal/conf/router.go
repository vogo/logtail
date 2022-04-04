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

package conf

import "github.com/vogo/logger"

type RouterConfigsFunc func() []*RouterConfig

func BuildRouterConfigsFunc(config *Config, serverConfig *ServerConfig) RouterConfigsFunc {
	return func() []*RouterConfig {
		var configs []*RouterConfig

		for _, name := range serverConfig.Routers {
			if r, ok := config.Routers[name]; ok {
				configs = append(configs, r)
			} else {
				logger.Errorf("router not exists: %s", name)
			}
		}

		return configs
	}
}
