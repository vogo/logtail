package logtail

import (
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
		routeToServerIndex(response, DefaultServerId)
		return
	}

	if strings.HasPrefix(uri, UriIndexPrefix+"/") {
		serverId := uri[len(UriIndexPrefix)+1:]
		if _, ok := serverDB[serverId]; ok {
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
		if _, ok := serverDB[tailServerId]; ok {
			response.WriteHeader(http.StatusNotFound)
			return
		}
	}

	if tailServerId != "" {
		startWebsocketTransfer(response, request, tailServerId)
		return
	}

	response.WriteHeader(http.StatusNotFound)
	return

}

func routeToServerIndex(response http.ResponseWriter, id string) {
	_, _ = response.Write(indexHTMLContent)
	return
}
