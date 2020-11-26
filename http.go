package logtail

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

var upgrader = websocket.Upgrader{}
var connCount int64 = 0

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

	connIndex := atomic.AddInt64(&connCount, 1)

	w := &wsWriter{
		ch:   make(chan struct{}),
		conn: c,
	}

	l.writer.AddWriter(w)

	logger.Infof("new connection %d", connIndex)

	go func() {
		defer func() {
			_ = recover()
		}()
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-w.ch:
				break
			case <-ticker.C:
				if _, _, err := c.ReadMessage(); err != nil {
					_ = w.Close()
					break
				}
			}
		}
	}()

	<-w.ch

	logger.Infof("connection %d closed", connIndex)
}
