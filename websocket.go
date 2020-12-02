package logtail

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

type websocketTransfer struct {
	transfer
	once sync.Once
	ch   chan struct{}
	conn *websocket.Conn
}

func NewWebsocketTransfer(index int64, c *websocket.Conn) *websocketTransfer {
	wt := &websocketTransfer{
		transfer: transfer{
			index:     index,
			transChan: make(chan []byte, 10),
		},
		once: sync.Once{},
		ch:   make(chan struct{}),
		conn: c,
	}
	return wt
}

func (wt *websocketTransfer) Close() {
	wt.once.Do(func() {
		logger.Infof("ws connection %d closing", wt.index)
		close(wt.ch)
	})
}

func (wt *websocketTransfer) Start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("ws connection %d error: %+v", wt.index, err)
		}
		logger.Infof("ws connection %d stopped", wt.index)
	}()

	logger.Infof("ws connection %d start", wt.index)

	defaultLogtailWriter.addTransfer(&wt.transfer)
	defer defaultLogtailWriter.removeTransfer(&wt.transfer)

	go func() {
		defer func() {
			_ = recover()
			wt.Close()
			logger.Warnf("ws connection %d heartbeat stopped", wt.index)
		}()

		for {
			select {
			case <-wt.ch:
				return
			default:
				_ = wt.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
				if _, _, err := wt.conn.ReadMessage(); err != nil {
					logger.Warnf("ws connection %d heartbeat error: %+v", wt.index, err)
					wt.Close()
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-wt.ch:
			return
		case <-ticker.C:
			continue
		case bytes := <-wt.transChan:
			if bytes == nil {
				wt.Close()
				return
			}
			if err := wt.conn.WriteMessage(1, bytes); err != nil {
				logger.Warnf("ws connection %d write error: %+v", wt.index, err)
				wt.Close()
				return
			}
		}

	}
}
