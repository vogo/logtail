/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	IDS
	conn *websocket.Conn
}

func (ww *WebsocketTransfer) start() error { return nil }

func (ww *WebsocketTransfer) stop() error { return nil }

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

	server, ok := defaultRunner.Servers[serverID]
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
		case <-router.stopper.C:
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
				if configErr := handleMatcherConfigUpdate(router, data[1:]); configErr != nil {
					logger.Warnf("router [%s] websocket matcher config error: %+v", router.name, configErr)
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
