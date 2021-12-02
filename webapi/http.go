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
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/vogo/logtail"
)

const (
	// URIIndexPrefix uri index prefix.
	URIIndexPrefix = "/index"

	// URITailPrefix uri tail prefix.
	URITailPrefix = "/tail"

	// URIManagePrefix uri manage prefix.
	URIManagePrefix = "/manage/"
)

type HTTPHandler struct {
	runner *logtail.Runner
}

// ServeHTTP serve http.
func (l *HTTPHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	Serve(request, response, l.runner)
}

func splitRouter(r string) (first, left string) {
	if i := strings.Index(r, "/"); i > 0 {
		return r[:i], r[i+1:]
	}

	return r, ""
}

// Serve web api
// routers:
// - `/index/<server-id>`: server index page
// - `/tail/<server-id>`: server tailing api
// - `/manage/<op>`: manage page
// - else route to default server list page.
func Serve(request *http.Request, response http.ResponseWriter, runner *logtail.Runner) {
	uri := request.RequestURI

	router, leftRouter := splitRouter(uri)
	switch router {
	case URIIndexPrefix:
		routeToIndexPage(runner, response, leftRouter)
	case URITailPrefix:
		routeToTail(runner, request, response, leftRouter)
	case URIManagePrefix:
		routeToManage(runner, request, response, leftRouter)
	default:
		responseServerList(runner, response)
	}
}

func routeToNotFound(response http.ResponseWriter) {
	response.WriteHeader(http.StatusNotFound)
}

func responseServerList(runner *logtail.Runner, response http.ResponseWriter) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(`<ul>`)

	for k := range runner.Servers {
		buf.WriteString(fmt.Sprintf(`<li><a href="/index/%s" target=_blank>%s</a></li>`, k, k))
	}

	buf.WriteString(`</ul>`)
	_, _ = response.Write(buf.Bytes())
}
