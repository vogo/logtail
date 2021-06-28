package logtail

import "sync"

// Stopper the stop status holder.
type Stopper struct {
	once sync.Once
	stop chan struct{}
}

// Stop stop the chan.
func (s *Stopper) Stop() {
	s.once.Do(func() {
		close(s.stop)
	})
}

// NewStopper create a new Stopper.
func NewStopper() *Stopper {
	return &Stopper{
		once: sync.Once{},
		stop: make(chan struct{}),
	}
}
