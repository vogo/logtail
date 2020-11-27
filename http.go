package logtail

import (
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

var upgrader = websocket.Upgrader{}

var connNum int64 = 0

type httpHandler struct {
}

func (l *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.RequestURI != "/tail" {
		_, _ = response.Write(indexHTMLContent)
		return
	}

	c, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		logger.Error("web socket error:", err)
		return
	}
	defer c.Close()

	connIndex := atomic.AddInt64(&connNum, 1)

	NewWebsocketTransfer(connIndex, c).Start()
}
