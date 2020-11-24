package logtail

import (
	"io"
	"sync"
)

type logtailWriter struct {
	lock    sync.Mutex
	writers []io.WriteCloser
}

func (ow *logtailWriter) AddWriter(w io.WriteCloser) {
	ow.lock.Lock()
	defer ow.lock.Unlock()

	ow.writers = append(ow.writers, w)
}

func (ow *logtailWriter) Write(p []byte) (n int, err error) {
	ow.lock.Lock()
	defer ow.lock.Unlock()

	if len(ow.writers) == 0 {
		return
	}

	var fails []int
	for idx, w := range ow.writers {
		if _, err := w.Write(p); err != nil {
			fails = append(fails, idx)
			_ = w.Close()
		}
	}

	if len(fails) > 0 {
		var filterWriters []io.WriteCloser
		from := 0
		for _, i := range fails {
			if i > from {
				filterWriters = append(filterWriters, ow.writers[from:i]...)
			}
			from = i
		}
		ow.writers = filterWriters
	}

	return len(p), nil
}
