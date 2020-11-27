package logtail

import (
	"sync"
)

type logtailWriter struct {
	lock    sync.Mutex
	writers map[int64]*websocketTransfer
}

func (ltw *logtailWriter) addTransfer(wt *websocketTransfer) {
	ltw.lock.Lock()
	defer ltw.lock.Unlock()

	ltw.writers[wt.index] = wt
}

func (ltw *logtailWriter) removeTransfer(wt *websocketTransfer) {
	ltw.lock.Lock()
	defer ltw.lock.Unlock()

	delete(ltw.writers, wt.index)
}

func (ltw *logtailWriter) Write(bytes []byte) (int, error) {
	ltw.lock.Lock()
	defer ltw.lock.Unlock()

	if len(ltw.writers) == 0 {
		return len(bytes), nil
	}

	for _, wt := range ltw.writers {
		wt.transferChan <- bytes
	}

	return len(bytes), nil
}
