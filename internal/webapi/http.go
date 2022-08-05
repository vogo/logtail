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

	"github.com/vogo/logtail/internal/tail"
)

const (
	// URIRouterIndex uri index router.
	URIRouterIndex = "index"

	// URIRouterTail uri tail router.
	URIRouterTail = "tail"

	// URIRouterManage uri manage router.
	URIRouterManage = "manage"
)

type HTTPHandler struct {
	runner *tail.Tailer
}

// ServeHTTP serve http.
func (l *HTTPHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	Serve(request, response, l.runner)
}

// split with / into two part.
//nolint:nonamedreturns //ignore this.
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
func Serve(request *http.Request, response http.ResponseWriter, runner *tail.Tailer) {
	uri := request.RequestURI
	if uri[0] == '/' {
		uri = uri[1:]
	}

	router, leftRouter := splitRouter(uri)
	switch router {
	case URIRouterTail:
		routeToTail(runner, request, response, leftRouter)
	case URIRouterManage:
		routeToManage(runner, request, response, leftRouter)
	case URIRouterIndex:
		routeToIndex(runner, request, response, leftRouter)
	default:
		responseServerList(runner, response)
	}
}

func responseServerList(runner *tail.Tailer, response http.ResponseWriter) {
	buf := bytes.NewBuffer(nil)

	buf.WriteString(`<ul>`)

	for k := range runner.Servers {
		buf.WriteString(fmt.Sprintf(`<li><a href="/index/%s" target=_blank>%s</a></li>`, k, k))
	}

	buf.WriteString(`</ul>`)
	response.Header().Add("Content-Type", "text/html")
	_, _ = response.Write(buf.Bytes())
}
