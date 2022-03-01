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

package webapi

import (
	"net/http"

	"github.com/vogo/logtail/internal/tailer"
)

func routeToManage(runner *tailer.Runner, request *http.Request, response http.ResponseWriter, router string) {
	firstRouter, leftRouter := splitRouter(router)
	switch firstRouter {
	case "index":
		routeToManageIndex(runner, request, response, leftRouter)
	case "transfer":
		routeToTransfer(runner, request, response, leftRouter)
	case "router":
		routeToRouter(runner, request, response, leftRouter)
	case "server":
		routeToServer(runner, request, response, leftRouter)
	default:
		routeToNotFound(response)
	}
}

// nolint:interfacer // ignore this
func routeToManageIndex(_ *tailer.Runner, _ *http.Request, response http.ResponseWriter, _ string) {
	_, _ = response.Write(manageHTMLContent)
}
