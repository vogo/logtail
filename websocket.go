package logtail

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

const WebsocketHeartbeatReadTimeout = 15 * time.Second

type WebsocketTransfer struct {
	conn *websocket.Conn
}

func (ww *WebsocketTransfer) start(*Router) error { return nil }

func (ww *WebsocketTransfer) Trans(_ string, data ...[]byte) (err error) {
	for _, d := range data {
		err = ww.conn.WriteMessage(1, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func startWebsocketTransfer(response http.ResponseWriter, request *http.Request, serverID string) {
	c, err := websocketUpgrader.Upgrade(response, request, nil)
	if err != nil {
		logger.Error("web socket error:", err)

		return
	}
	defer c.Close()

	server, ok := serverDB[serverID]
	if !ok {
		logger.Warnf("server id not found: %s", serverID)

		return
	}

	websocketTransfer := &WebsocketTransfer{conn: c}
	router := NewRouter(server, nil, []Transfer{websocketTransfer})
	server.mergingWorker.startRouterFilter(router)
	startWebsocketHeartbeat(router, websocketTransfer)
}

const MessageTypeMatcherConfig = '1'

func startWebsocketHeartbeat(router *Router, transfer *WebsocketTransfer) {
	defer func() {
		_ = recover()

		router.Stop()
		logger.Infof("router [%s] websocket heartbeat stopped", router.name)
	}()

	for {
		select {
		case <-router.close:
			return
		default:
			_ = transfer.conn.SetReadDeadline(time.Now().Add(WebsocketHeartbeatReadTimeout))

			_, data, err := transfer.conn.ReadMessage()
			if err != nil {
				if !isEncodeError(err) {
					logger.Warnf("router [%s] websocket heartbeat error: %+v", router.name, err)
					router.Stop()
				}

				return
			}

			if len(data) > 0 && data[0] == MessageTypeMatcherConfig {
				if err := handleMatcherConfigUpdate(router, data[1:]); err != nil {
					logger.Warnf("router [%s] websocket matcher config error: %+v", router.name, err)
				}
			}
		}
	}
}

func isEncodeError(err error) bool {
	return strings.Contains(err.Error(), "utf8")
}

func handleMatcherConfigUpdate(router *Router, data []byte) error {
	var matcherConfigs []*MatcherConfig
	if err := json.Unmarshal(data, &matcherConfigs); err != nil {
		return err
	}

	if err := validateMatchers(matcherConfigs); err != nil {
		return err
	}

	router.SetMatchers(buildMatchers(matcherConfigs))

	return nil
}
