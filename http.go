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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/vogo/logger"
)

type httpHandler struct{}

const (
	// URIIndexPrefix uri index prefix.
	URIIndexPrefix = "/index"

	// URITailPrefix uri tail prefix.
	URITailPrefix = "/tail"

	// URIManagePrefix uri manage prefix.
	URIManagePrefix = "/manage/"
)

// ServeHTTP serve http
// routers:
// /: default server index page
// /tail: default server tailing api
// /index/<server-id>: server index page
// /tail/<server-id>: server tailing api
// /manage: manage page.
func (l *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	uri := request.RequestURI

	if strings.HasPrefix(uri, URIIndexPrefix+"/") {
		routeToIndexPage(response, uri)

		return
	}

	if strings.HasPrefix(uri, URITailPrefix) {
		routeToTail(request, response, uri)

		return
	}

	if strings.HasPrefix(uri, URIManagePrefix) {
		routeToManage(request, response, uri[len(URIManagePrefix)+1:])

		return
	}

	responseServerList(response)
}

func routeToManage(_ *http.Request, response http.ResponseWriter, op string) {
	switch op {
	case "transfer/list":
		listTransfers(response)
	case "transfer/add":
	case "transfer/start":
	case "transfer/stop":
		logger.Info(op)
	}
}

func listTransfers(response http.ResponseWriter) {
	response.Header().Add("content-type", "application/json")

	b, _ := json.Marshal(defaultRunner.Config.Transfers)

	_, _ = response.Write(b)
}

func routeToTail(request *http.Request, response http.ResponseWriter, uri string) {
	tailServerID := ""
	if uri == URITailPrefix {
		tailServerID = DefaultID
	} else if strings.HasPrefix(uri, URITailPrefix+"/") {
		tailServerID = uri[len(URITailPrefix)+1:]
		if _, ok := defaultRunner.Servers[tailServerID]; !ok {
			tailServerID = ""
		}
	}

	if tailServerID == "" {
		response.WriteHeader(http.StatusNotFound)

		return
	}

	startWebsocketTransfer(response, request, tailServerID)
}

func responseServerList(response http.ResponseWriter) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(`<ul>`)

	for k := range defaultRunner.Servers {
		buf.WriteString(fmt.Sprintf(`<li><a href="/index/%s" target=_blank>%s</a></li>`, k, k))
	}

	buf.WriteString(`</ul>`)
	_, _ = response.Write(buf.Bytes())
}

func routeToIndexPage(response http.ResponseWriter, uri string) {
	serverID := uri[len(URIIndexPrefix)+1:]
	_, ok := defaultRunner.Servers[serverID]

	if !ok {
		response.WriteHeader(http.StatusNotFound)

		return
	}

	_, _ = response.Write(indexHTMLContent)
}
