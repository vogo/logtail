package logtail

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/vogo/logger"
)

var ErrWorkerCommandStopped = errors.New("worker command stopped")

type worker struct {
	id      string
	server  *Server
	dynamic bool      // command generated dynamically
	command string    // command lines
	cmd     *exec.Cmd // command object
	filters map[int64]*Filter
}

func (w *worker) Write(data []byte) (int, error) {
	// copy data to avoid being update by source
	d := make([]byte, len(data))
	copy(d, data)

	for _, r := range w.filters {
		r.receive(d)
	}

	_, _ = w.server.Write(d)

	return len(d), nil
}

func (w *worker) writeToFilters(bytes []byte) (int, error) {
	for _, r := range w.filters {
		r.receive(bytes)
	}

	return len(bytes), nil
}

func (w *worker) startRouterFilter(router *Router) {
	select {
	case <-w.server.stop:
		return
	default:
		filter := newFilter(w, router)
		w.filters[router.id] = filter

		go func() {
			defer delete(w.filters, router.id)
			filter.start()
		}()
	}
}

func (w *worker) start() {
	if w.command == "" {
		return
	}

	go func() {
		defer func() {
			w.stop()
			logger.Infof("worker [%s] stopped", w.id)
		}()

		for {
			select {
			case <-w.server.stop:
				return
			default:
				logger.Infof("worker [%s] command: %s", w.id, w.command)
				w.cmd = exec.Command("/bin/sh", "-c", w.command)

				setCmdSysProcAttr(w.cmd)

				w.cmd.Stdout = w
				w.cmd.Stderr = os.Stderr

				if err := w.cmd.Run(); err != nil {
					logger.Errorf("worker [%s] command error: %+v, command: %s", w.id, err, w.command)

					// if the command is generated dynamic, should not restart by self, send error instead.
					if w.dynamic {
						w.server.receiveWorkerError(err)
						return
					}

					select {
					case <-w.server.stop:
						return
					default:
						logger.Errorf("worker [%s] failed, retry after 10s! command: %s", w.id, w.command)
						time.Sleep(CommandFailRetryInterval)
					}
				}

				// if the command is generated dynamic, should not restart by self, send error instead.
				if w.dynamic {
					w.server.workerError <- fmt.Errorf("%w: worker [%s]", ErrWorkerCommandStopped, w.id)
					return
				}
			}
		}
	}()
}

func (w *worker) stop() {
	defer func() {
		if err := recover(); err != nil {
			logger.Warnf("worker [%s] close error: %+v", w.id, err)
		}
	}()

	if w.cmd != nil {
		logger.Infof("worker [%s] command stopping: %s", w.id, w.command)

		if err := killCmd(w.cmd); err != nil {
			logger.Warnf("worker [%s] kill command error: %+v", w.id, err)
		}

		w.cmd = nil
	}

	w.stopFilters()
}

func (w *worker) stopFilters() {
	for _, filter := range w.filters {
		filter.stop()
	}
}

func startWorker(s *Server, command string, dynamic bool) *worker {
	w := newWorker(s, command, dynamic)

	if len(s.routers) > 0 {
		for _, r := range s.routers {
			w.startRouterFilter(r)
		}
	}

	w.start()

	return w
}

func newWorker(s *Server, command string, dynamic bool) *worker {
	id := fmt.Sprintf("%s-%d", s.id, len(s.workers))
	if command == "" {
		id = fmt.Sprintf("%s-default", s.id)
	}

	return &worker{
		id:      id,
		server:  s,
		command: command,
		dynamic: dynamic,
		filters: make(map[int64]*Filter, 4),
	}
}
