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

// FromStopper create a new Stopper from a exist one, when which is stopped the new will be stooped too.
func FromStopper(from *Stopper) *Stopper {
	s := &Stopper{
		once: sync.Once{},
		stop: make(chan struct{}),
	}

	go func() {
		select {
		case <-from.stop:
			s.Stop()
		case <-s.stop:
		}
	}()

	return s
}
