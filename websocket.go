package logtail

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

type websocketTransfer struct {
	index        int64
	once         sync.Once
	ch           chan struct{}
	transferChan chan []byte
	conn         *websocket.Conn
}

func NewWebsocketTransfer(index int64, c *websocket.Conn) *websocketTransfer {
	wt := &websocketTransfer{
		index:        index,
		once:         sync.Once{},
		ch:           make(chan struct{}),
		transferChan: make(chan []byte, 10),
		conn:         c,
	}
	return wt
}

func (wt *websocketTransfer) Close() {
	wt.once.Do(func() {
		logger.Infof("connection %d closing", wt.index)
		close(wt.ch)
	})
}

func (wt *websocketTransfer) Start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("connection %d error: %+v", wt.index, err)
		}
		logger.Infof("connection %d stopped", wt.index)
	}()

	logger.Infof("connection %d start", wt.index)

	defaultLogtailWriter.addTransfer(wt)
	defer defaultLogtailWriter.removeTransfer(wt)

	go func() {
		defer func() {
			_ = recover()
			wt.Close()
			logger.Errorf("connection %d heartbeat stopped", wt.index)
		}()

		for {
			select {
			case <-wt.ch:
				return
			default:
				_ = wt.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
				if _, _, err := wt.conn.ReadMessage(); err != nil {
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
		case bytes := <-wt.transferChan:
			if bytes == nil {
				wt.Close()
				return
			}
			if err := wt.conn.WriteMessage(1, bytes); err != nil {
				wt.Close()
				return
			}
		}

	}
}
