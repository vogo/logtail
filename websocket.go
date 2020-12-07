package logtail

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

type WebsocketTransfer struct {
	conn *websocket.Conn
}

func (ww *WebsocketTransfer) Trans(bytes []byte) error {
	return ww.conn.WriteMessage(1, bytes)
}

func startWebsocketHeartbeat(t *Router, c *websocket.Conn) {
	defer func() {
		_ = recover()
		t.Stop()
		logger.Infof("router %s websocket heartbeat stopped", t.id)
	}()

	for {
		select {
		case <-t.stop:
			return
		default:
			_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
			if _, _, err := c.ReadMessage(); err != nil {
				logger.Warnf("router %s websocket heartbeat error: %+v", t.id, err)
				t.Stop()
				return
			}
		}
	}
}

func startWebsocketTransfer(response http.ResponseWriter, request *http.Request, serverId string) {
	c, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		logger.Error("web socket error:", err)
		return
	}
	defer c.Close()

	server, ok := serverDB[serverId]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	websocketTransfer := &WebsocketTransfer{conn: c}
	router := NewRouter(nil, []Transfer{websocketTransfer})
	server.StartRouter(router)
	startWebsocketHeartbeat(router, c)
}
