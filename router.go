package logtail

import (
	"sync"

	"github.com/vogo/logger"
)

var defaultRouters []*Router
var globalRouters []*Router

type Router struct {
	id        string
	channel   Channel
	lock      sync.Mutex
	once      sync.Once
	stop      chan struct{}
	matchers  []Matcher
	transfers []Transfer
}

func NewRouter(id string, matchers []Matcher, transfers []Transfer) *Router {
	t := &Router{
		id:        id,
		channel:   make(chan *Message, 16),
		lock:      sync.Mutex{},
		once:      sync.Once{},
		stop:      make(chan struct{}),
		matchers:  matchers,
		transfers: transfers,
	}
	return t
}

func (r *Router) SetMatchers(matchers []Matcher) {
	r.matchers = matchers
}

func (r *Router) SetTransfer(transfers []Transfer) {
	r.transfers = transfers
}

func (r *Router) Route(serverId string, bytes []byte) {
	routeMatchers := r.matchers
	if len(routeMatchers) > 0 {
		if err := r.MatchAndTrans(serverId, routeMatchers, bytes); err != nil {
			logger.Warnf("router %s write error: %+v", r.id, err)
			r.Stop()
		}
		return
	}

	if err := r.Trans(serverId, bytes); err != nil {
		logger.Warnf("router %s write error: %+v", r.id, err)
		r.Stop()
	}
}

func (r *Router) MatchAndTrans(serverId string, matchers []Matcher, bytes []byte) error {
	matches := matchers[0].Match(bytes)
	if len(matches) == 0 {
		return nil
	}
	if len(matchers) == 1 {
		for _, data := range matches {
			if err := r.Trans(serverId, data); err != nil {
				return err
			}
		}
		return nil
	}

	for _, data := range matches {
		if err := r.MatchAndTrans(serverId, matchers[1:], data); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) Trans(serverId string, bytes []byte) error {
	transfers := r.transfers
	if len(transfers) == 0 {
		return nil
	}
	for _, t := range transfers {
		if err := t.Trans(serverId, bytes); err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) Stop() {
	r.once.Do(func() {
		logger.Infof("router %s closing", r.id)
		close(r.stop)
		close(r.channel)
	})
}

func (r *Router) Start() {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("router %s error: %+v", r.id, err)
		}
		logger.Infof("router %s stopped", r.id)
	}()

	logger.Infof("router %s start", r.id)

	for {
		select {
		case <-r.stop:
			return
		case message := <-r.channel:
			if message == nil {
				r.Stop()
				return
			}
			r.Route(message.ServerId, message.Data)
		}
	}
}

func (r *Router) receive(message *Message) {
	select {
	case <-r.stop:
		return
	case r.channel <- message:
	default:
	}
}
