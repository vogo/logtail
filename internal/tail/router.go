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

package tail

import (
	"fmt"

	"github.com/vogo/logger"
	"github.com/vogo/logtail/internal/conf"
)

func (t *Tailer) AddRouter(config *conf.RouterConfig) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := conf.CheckRouterConfig(t.Config, config); err != nil {
		return err
	}

	var err error

	if _, ok := t.Config.Routers[config.Name]; ok {
		for _, server := range t.Servers {
			for _, worker := range server.Workers {
				for _, router := range worker.Routers {
					if router.Name == config.Name {
						if err = worker.AddRouter(config); err != nil {
							logger.Errorf("add Routers error: %v", err)
						}
					}
				}
			}
		}
	}

	t.Config.Routers[config.Name] = config
	t.Config.SaveToFile()

	return err
}

func (t *Tailer) DeleteRouter(name string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if _, exist := t.Config.Routers[name]; exist {
		if t.IsRouterUsing(name) {
			return fmt.Errorf("%w: %s", conf.ErrRouterUsing, name)
		}

		delete(t.Config.Routers, name)
		t.Config.SaveToFile()
	}

	return nil
}

func (t *Tailer) IsRouterUsing(name string) bool {
	for _, server := range t.Config.Servers {
		for _, router := range server.Routers {
			if router == name {
				return true
			}
		}
	}

	return false
}
