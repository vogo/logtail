package logtail

import (
	"sync"
)

type logtailWriter struct {
	lock    sync.Mutex
	writers map[int64]*transfer
}

func (ltw *logtailWriter) addTransfer(wt *transfer) {
	ltw.lock.Lock()
	defer ltw.lock.Unlock()

	ltw.writers[wt.index] = wt
}

func (ltw *logtailWriter) removeTransfer(wt *transfer) {
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
		wt.transChan <- bytes
	}

	return len(bytes), nil
}
