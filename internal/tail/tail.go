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
	"sync"
	"time"

	"github.com/vogo/logtail/internal/conf"
	"github.com/vogo/logtail/internal/serve"
	"github.com/vogo/logtail/internal/trans"
	"github.com/vogo/logtail/internal/util"
	"github.com/vogo/vogo/vlog"
)

// Tailer the logtail tailer.
type Tailer struct {
	lock      sync.Mutex
	Config    *conf.Config
	Servers   map[string]*serve.Server
	Transfers map[string]trans.Transfer
}

// NewTailer new logtail tailer.
func NewTailer(config *conf.Config) (*Tailer, error) {
	if err := conf.InitialCheckConfig(config); err != nil {
		return nil, err
	}

	tailer := &Tailer{
		lock:      sync.Mutex{},
		Config:    config,
		Servers:   make(map[string]*serve.Server, util.DefaultMapSize),
		Transfers: make(map[string]trans.Transfer, util.DefaultMapSize),
	}

	return tailer, nil
}

func (t *Tailer) Start() error {
	conf.ConfigLogLevel(t.Config.LogLevel)

	if t.Config.StatisticPeriodMinutes > 0 {
		trans.SetTransStatisticDuration(time.Duration(t.Config.StatisticPeriodMinutes) * time.Minute)
	}

	if err := t.StartTransfers(); err != nil {
		return err
	}

	for _, serverConfig := range t.Config.Servers {
		_, err := t.AddServer(serverConfig)
		if err != nil {
			vlog.Errorf("add server %s error: %v", serverConfig.Name, err)
		}
	}

	return nil
}

// RouterStats holds pipeline statistics for a single router.
type RouterStats struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Source       string `json:"source"`
	DropCount    int64  `json:"drop_count"`
	BufferSize   int    `json:"buffer_size"`
	BlockingMode bool   `json:"blocking_mode"`
}

// CollectRouterStats returns pipeline statistics for all active routers.
func (t *Tailer) CollectRouterStats() []RouterStats {
	t.lock.Lock()
	defer t.lock.Unlock()

	stats := make([]RouterStats, 0)

	for _, server := range t.Servers {
		for _, worker := range server.Workers {
			for _, router := range worker.Routers {
				stats = append(stats, RouterStats{
					ID:           router.ID,
					Name:         router.Name,
					Source:       router.Source,
					DropCount:    router.DroppedMessages(),
					BufferSize:   router.BufferSize,
					BlockingMode: router.BlockingMode,
				})
			}
		}
	}

	return stats
}

// Stop the runner.
func (t *Tailer) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, s := range t.Servers {
		if err := s.Stop(); err != nil {
			vlog.Errorf("server %s close error: %+v", s.ID, err)
		}
	}

	for _, t := range t.Transfers {
		if err := t.Stop(); err != nil {
			vlog.Errorf("transfer %s close error: %+v", t.Name(), err)
		}
	}
}
