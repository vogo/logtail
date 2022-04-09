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

package work

import (
	"fmt"

	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/route"
)

func (w *Worker) AddRouter(routerConfig *conf.RouterConfig) error {
	routerName := routerConfig.Name
	if existRouter, exist := w.Routers[routerName]; exist {
		existRouter.Stop()
	}

	routerID := fmt.Sprintf("%s-%s", w.ID, routerName)

	router := route.StartRouter(w.Runner, routerConfig, w.TransfersFunc, routerID, w.Source)

	w.Routers[routerName] = router

	return nil
}

func (w *Worker) WriteToRouters(bytes []byte) (int, error) {
	for _, r := range w.Routers {
		r.Receive(bytes)
	}

	return len(bytes), nil
}

func (w *Worker) StartRouter(router *route.Router) {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.Runner.C:
		return
	default:
		w.Routers[router.ID] = router

		go func() {
			defer delete(w.Routers, router.ID)
			router.StartLoop()
		}()
	}
}
