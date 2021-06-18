package logtail

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type httpHandler struct{}

const (
	// URIIndexPrefix uri index prefix.
	URIIndexPrefix = "/index"

	// URITailPrefix uri tail prefix.
	URITailPrefix = "/tail"
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

	if uri == "" || uri == "/" || uri == URIIndexPrefix {
		responseServerList(response)

		return
	}

	if strings.HasPrefix(uri, URIIndexPrefix+"/") {
		serverID := uri[len(URIIndexPrefix)+1:]
		_, ok := serverDB[serverID]

		if !ok {
			response.WriteHeader(http.StatusNotFound)

			return
		}

		routeToServerIndex(response, serverID)

		return
	}

	tailServerID := ""
	if uri == URITailPrefix {
		tailServerID = DefaultServerID
	} else if strings.HasPrefix(uri, URITailPrefix+"/") {
		tailServerID = uri[len(URITailPrefix)+1:]
		if _, ok := serverDB[tailServerID]; !ok {
			response.WriteHeader(http.StatusNotFound)

			return
		}
	}

	if tailServerID != "" {
		startWebsocketTransfer(response, request, tailServerID)

		return
	}

	responseServerList(response)
}

func responseServerList(response http.ResponseWriter) {
	buf := bytes.NewBuffer(nil)

	for k := range serverDB {
		buf.WriteString(fmt.Sprintf("<a href=\"/index/%s\" target=_blank>%s</a> ", k, k))
	}

	_, _ = response.Write(buf.Bytes())
}

func routeToServerIndex(response io.Writer, _ string) {
	_, _ = response.Write(indexHTMLContent)
}
