package logtail

import (
	"github.com/gorilla/websocket"
	"github.com/vogo/logger"
)

type wsWriter struct {
	ch   chan struct{}
	conn *websocket.Conn
}

func (ww *wsWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	if err = ww.conn.WriteMessage(1, p); err != nil {
		logger.Error("failed to write message!", err)
		return 0, err
	}
	return n, nil
}

func (ww *wsWriter) Close() error {
	defer func() {
		_ = recover()
	}()
	close(ww.ch)
	return nil
}
