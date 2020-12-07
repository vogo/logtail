package logtail

import (
	"encoding/json"
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

const MessageTypeConfigMatcher = '1'

func startWebsocketHeartbeat(router *Router, transfer *WebsocketTransfer) {
	defer func() {
		_ = recover()
		router.Stop()
		logger.Infof("router %s websocket heartbeat stopped", router.id)
	}()

	for {
		select {
		case <-router.stop:
			return
		default:
			_ = transfer.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			_, data, err := transfer.conn.ReadMessage()
			if err != nil {
				logger.Warnf("router %s websocket heartbeat error: %+v", router.id, err)
				router.Stop()
				return
			}

			if len(data) > 0 && data[0] == MessageTypeConfigMatcher {
				if err := handleRouterConfig(router, transfer, data[1:]); err != nil {
					logger.Warnf("router %s websocket router config error: %+v", router.id, err)
				}
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
	startWebsocketHeartbeat(router, websocketTransfer)
}

func handleRouterConfig(router *Router, transfer *WebsocketTransfer, data []byte) error {
	var routeConfig RouterConfig
	if err := json.Unmarshal(data, &routeConfig); err != nil {
		return err
	}
	if err := validateRouterConfig(&routeConfig); err != nil {
		return err
	}

	router.SetMatchers(buildMatchers(routeConfig.Matchers))
	transfers := buildTransfers(routeConfig.Transfers)
	transfers = append(transfers, transfer)
	router.SetTransfer(transfers)

	return nil
}
