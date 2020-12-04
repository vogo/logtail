package logtail

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

type websocketWriter struct {
	conn *websocket.Conn
}

func (ww *websocketWriter) Write(bytes []byte) (int, error) {
	return len(bytes), ww.conn.WriteMessage(1, bytes)
}

func startHeartbeat(t *Transfer, c *websocket.Conn) {
	defer func() {
		_ = recover()
		t.Close()
		logger.Warnf("ws connection %d heartbeat stopped", t.index)
	}()

	for {
		select {
		case <-t.ch:
			return
		default:
			_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
			if _, _, err := c.ReadMessage(); err != nil {
				logger.Warnf("ws connection %d heartbeat error: %+v", t.index, err)
				t.Close()
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{}

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

	t := NewTransfer(&websocketWriter{
		conn: c,
	})

	t.Start()

	startHeartbeat(t, c)
}

func startHttpListener() {
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), &httpHandler{}); err != nil {
		panic(err)
	}
}
