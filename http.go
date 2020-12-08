package logtail

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type httpHandler struct {
}

const UriIndexPrefix = "/index"
const UriTailPrefix = "/tail"

// default server index page: /
// default server tailing api: /tail
// server index page: /index/<server-id>
// server tailing api: /tail/<server-id>
// manage page: /manage
func (l *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	uri := request.RequestURI

	if uri == "" || uri == "/" || uri == UriIndexPrefix {
		responseServerList(response)
		return
	}

	if strings.HasPrefix(uri, UriIndexPrefix+"/") {
		serverId := uri[len(UriIndexPrefix)+1:]
		_, ok := serverDB[serverId]
		if !ok {
			response.WriteHeader(http.StatusNotFound)
			return
		}

		routeToServerIndex(response, serverId)
		return
	}

	tailServerId := ""
	if uri == UriTailPrefix {
		tailServerId = DefaultServerId
	} else if strings.HasPrefix(uri, UriTailPrefix+"/") {
		tailServerId = uri[len(UriTailPrefix)+1:]
		if _, ok := serverDB[tailServerId]; !ok {
			response.WriteHeader(http.StatusNotFound)
			return
		}
	}

	if tailServerId != "" {
		startWebsocketTransfer(response, request, tailServerId)
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

func routeToServerIndex(response http.ResponseWriter, id string) {
	_, _ = response.Write(indexHTMLContent)
}
