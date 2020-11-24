package logtail

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

var upgrader = websocket.Upgrader{}

type httpHandler struct {
	writer *logtailWriter
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

	w := &wsWriter{
		ch:   make(chan struct{}),
		conn: c,
	}
	l.writer.AddWriter(w)
	<-w.ch
}
