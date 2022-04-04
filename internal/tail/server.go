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

package tail

import (
	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/serve"
)

func (t *Tailer) AddServer(serverConfig *conf.ServerConfig) (*serve.Server, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if err := conf.CheckServerConfig(t.Config, serverConfig); err != nil {
		return nil, err
	}

	server := buildServer(serverConfig, t)

	server.Start(serverConfig)

	t.Config.Servers[serverConfig.Name] = serverConfig

	t.Config.SaveToFile()

	return server, nil
}

func (t *Tailer) DeleteServer(name string) error {
	s, exist := t.Servers[name]

	if exist {
		if err := s.Stop(); err != nil {
			return err
		}

		delete(t.Servers, name)

		delete(t.Config.Servers, name)
		t.Config.SaveToFile()
	}

	return nil
}

func buildServer(serverConfig *conf.ServerConfig, tailer *Tailer) *serve.Server {
	server := serve.NewRawServer(serverConfig.Name)

	format := serverConfig.Format
	if format == nil {
		format = tailer.Config.DefaultFormat
	}

	server.Format = format

	if existsServer, ok := tailer.Servers[server.ID]; ok {
		_ = existsServer.Stop()

		delete(tailer.Servers, server.ID)
	}

	tailer.Servers[server.ID] = server

	server.TransferMatcher = buildTransferMatcher(tailer)
	server.RouterConfigsFunc = conf.BuildRouterConfigsFunc(tailer.Config, serverConfig)

	return server
}
