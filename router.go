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
	stop      chan struct{}
	matchers  []Matcher
	transfers []Transfer
}

func NewRouter(id string, matchers []Matcher, transfers []Transfer) *Router {
	t := &Router{
		id:        id,
		channel:   make(chan *Message, DefaultChannelBufferSize),
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

func (r *Router) Route(server *Server, bytes []byte) {
	if err := r.MatchAndTrans(server, r.matchers, bytes); err != nil {
		logger.Warnf("router %s route error: %+v", r.id, err)
		r.Stop()
	}
}

func (r *Router) MatchAndTrans(server *Server, matchers []Matcher, bytes []byte) error {
	if len(matchers) == 0 {
		return r.Trans(server.id, bytes)
	}

	matches := matchers[0].Match(server.format, bytes)

	if len(matches) == 0 {
		return nil
	}

	matchers = matchers[1:]

	for _, data := range matches {
		if err := r.MatchAndTrans(server, matchers, data); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) Trans(serverID string, bytes []byte) error {
	transfers := r.transfers
	if len(transfers) == 0 {
		return nil
	}

	for _, t := range transfers {
		if err := t.Trans(serverID, bytes); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) Stop() {
	r.once.Do(func() {
		logger.Infof("router %s stopping", r.id)
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

			r.Route(message.Server, message.Data)
		}
	}
}

func (r *Router) receive(message *Message) {
	defer func() {
		_ = recover()
	}()

	select {
	case <-r.stop:
		return
	case r.channel <- message:
	default:
	}
}
