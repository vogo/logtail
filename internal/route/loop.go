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

package route

import (
	"runtime/debug"

	"github.com/vogo/logger"
)

func (r *Router) StartLoop() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Routers [%s] error: %+v, stack:\n%s", r.ID, err, string(debug.Stack()))
		}

		logger.Infof("Routers [%s] stopped", r.ID)
	}()

	logger.Infof("Routers [%s] StartLoop", r.ID)

	for {
		select {
		case <-r.Runner.C:
			return
		case data := <-r.channel:
			if data == nil {
				r.Stop()

				return
			}

			if err := r.Route(data); err != nil {
				logger.Warnf("Routers [%s] route error: %+v", r.ID, err)
				r.Stop()
			}
		}
	}
}
