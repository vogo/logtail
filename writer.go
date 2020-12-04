package logtail

import (
	"sync"
)

var defaultLogtailWriter = &logtailWriter{
	lock:    sync.Mutex{},
	tunnels: make(map[int64]Tunnel, 16),
}

type logtailWriter struct {
	lock    sync.Mutex
	tunnels map[int64]Tunnel
}

func (ltw *logtailWriter) addTunnel(index int64, t Tunnel) {
	ltw.lock.Lock()
	defer ltw.lock.Unlock()

	ltw.tunnels[index] = t
}

func (ltw *logtailWriter) removeTunnel(index int64) {
	ltw.lock.Lock()
	defer ltw.lock.Unlock()

	delete(ltw.tunnels, index)
}

func (ltw *logtailWriter) Write(bytes []byte) (int, error) {
	ltw.lock.Lock()
	defer ltw.lock.Unlock()

	if len(ltw.tunnels) == 0 {
		return len(bytes), nil
	}

	for _, t := range ltw.tunnels {
		select {
		case t <- bytes:
		default:
		}
	}

	return len(bytes), nil
}
