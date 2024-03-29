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
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/tail"
)

var ErrWebAPICommandNotSupported = errors.New("web api not support command source")

func routeToServer(runner *tail.Tailer, request *http.Request, response http.ResponseWriter, router string) {
	switch router {
	case OpTypes:
		listServerTypes(runner, response)
	case OpList:
		listServers(runner, response)
	case OpAdd:
		addServer(runner, request, response)
	case OpDelete:
		deleteServer(runner, request, response)
	default:
		routeToNotFound(response)
	}
}

func listServerTypes(_ *tail.Tailer, response http.ResponseWriter) {
	response.Header().Add("content-type", "application/json")

	//nolint:errchkjson //ignore this
	b, _ := json.Marshal(conf.ServerTypes)

	_, _ = response.Write(b)
}

func listServers(runner *tail.Tailer, response http.ResponseWriter) {
	response.Header().Add("content-type", "application/json")

	//nolint:errchkjson //ignore this
	b, _ := json.Marshal(runner.Config.Servers)

	_, _ = response.Write(b)
}

func addServer(runner *tail.Tailer, request *http.Request, response http.ResponseWriter) {
	config := &conf.ServerConfig{}

	if err := json.NewDecoder(request.Body).Decode(config); err != nil {
		routeToError(response, err)

		return
	}

	if config.Command != "" || config.Commands != "" || config.CommandGen != "" {
		routeToError(response, ErrWebAPICommandNotSupported)

		return
	}

	if _, err := runner.AddServer(config); err != nil {
		routeToError(response, err)

		return
	}

	routeToSuccess(response)
}

func deleteServer(runner *tail.Tailer, request *http.Request, response http.ResponseWriter) {
	config := &conf.ServerConfig{}

	if err := json.NewDecoder(request.Body).Decode(config); err != nil {
		routeToError(response, err)

		return
	}

	if err := runner.DeleteServer(config.Name); err != nil {
		routeToError(response, err)

		return
	}

	routeToSuccess(response)
}
