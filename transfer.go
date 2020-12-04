package logtail

import (
	"io"
	"sync"

	"github.com/vogo/logger"
)

type Transfer struct {
	tunnel   Tunnel
	index    int64
	once     sync.Once
	ch       chan struct{}
	consumer io.Writer
}

func NewTransfer(c io.Writer) *Transfer {
	t := &Transfer{
		tunnel:   make(chan []byte, 16),
		once:     sync.Once{},
		ch:       make(chan struct{}),
		consumer: c,
	}
	return t
}

func (t *Transfer) Close() {
	t.once.Do(func() {
		logger.Infof("transfer %d closing", t.index)
		close(t.ch)
	})
}

func (t *Transfer) Start() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("transfer %d error: %+v", t.index, err)
			}
			logger.Infof("transfer %d stopped", t.index)
		}()

		t.index = AddTunnel(t.tunnel)
		defer RemoveTunnel(t.index)

		logger.Infof("transfer %d start", t.index)

		for {
			select {
			case <-t.ch:
				return
			case bytes := <-t.tunnel:
				if bytes == nil {
					t.Close()
					return
				}
				if _, err := t.consumer.Write(bytes); err != nil {
					logger.Warnf("transfer %d write error: %+v", t.index, err)
					t.Close()
					return
				}
			}
		}
	}()
}
