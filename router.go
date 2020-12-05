package logtail

import (
	"sync"

	"github.com/vogo/logger"
)

type Router struct {
	id        string
	channel   Channel
	lock      sync.Mutex
	once      sync.Once
	ch        chan struct{}
	filters   []*Filter
	transfers []Transfer
}

func NewRouter(filters []*Filter, transfers []Transfer) *Router {
	t := &Router{
		channel:   make(chan []byte, 16),
		lock:      sync.Mutex{},
		once:      sync.Once{},
		ch:        make(chan struct{}),
		filters:   filters,
		transfers: transfers,
	}
	return t
}

func (r *Router) SetFilters(filters []*Filter) {
	r.filters = filters
}

func (r *Router) SetTransfer(transfers []Transfer) {
	r.transfers = transfers
}

func (r *Router) Close() {
	r.once.Do(func() {
		logger.Infof("transfer %s closing", r.id)
		close(r.ch)
	})
}

func (r *Router) Start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("transfer %s error: %+v", r.id, err)
		}
		logger.Infof("transfer %s stopped", r.id)
	}()

	logger.Infof("transfer %s start", r.id)

	for {
		select {
		case <-r.ch:
			return
		case bytes := <-r.channel:
			if bytes == nil {
				r.Close()
				return
			}
			r.Route(bytes)
		}
	}
}

func (r *Router) Route(bytes []byte) {
	routeFilters := r.filters
	if len(routeFilters) > 0 {
		if err := r.FilterAndTrans(bytes, routeFilters); err != nil {
			logger.Warnf("router %s write error: %+v", r.id, err)
			r.Close()
		}
		return
	}

	if err := r.Trans(bytes); err != nil {
		logger.Warnf("router %s write error: %+v", r.id, err)
		r.Close()
	}
}

func (r *Router) FilterAndTrans(bytes []byte, filters []*Filter) error {
	matches := filters[0].Matcher.Match(bytes)
	if len(matches) == 0 {
		return nil
	}
	if len(filters) == 1 {
		for _, data := range matches {
			if err := r.Trans(data); err != nil {
				return err
			}
		}
		return nil
	}

	for _, data := range matches {
		if err := r.FilterAndTrans(data, filters[1:]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) Trans(bytes []byte) error {
	transfers := r.transfers
	if len(transfers) == 0 {
		return nil
	}
	for _, t := range transfers {
		if err := t.Trans(bytes); err != nil {
			return err
		}
	}
	return nil
}
