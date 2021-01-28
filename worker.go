package logtail

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
)

var ErrWorkerCommandStopped = errors.New("worker command stopped")

type worker struct {
	id          string
	server      *Server
	dynamic     bool      // command generated dynamically
	command     string    // command lines
	cmd         *exec.Cmd // command object
	routerCount int64
	routers     map[int64]*Router
}

func (w *worker) Write(data []byte) (int, error) {
	// copy data to avoid being update by source
	d := make([]byte, len(data))
	copy(d, data)

	for _, r := range w.routers {
		r.receive(d)
	}

	_, _ = w.server.Write(d)

	return len(d), nil
}

func (w *worker) writeToRouter(bytes []byte) (int, error) {
	for _, r := range w.routers {
		r.receive(bytes)
	}

	return len(bytes), nil
}

func (w *worker) addRouter(router *Router) {
	router.worker = w

	select {
	case <-w.server.stop:
		return
	default:
		index := atomic.AddInt64(&w.routerCount, 1)

		if router.id == "" {
			router.id = fmt.Sprintf("%s-%d", w.id, index)
		}

		w.routers[index] = router

		go func() {
			defer delete(w.routers, index)
			router.start()
		}()
	}
}

func (w *worker) start() {
	if w.command == "" {
		return
	}

	go func() {
		defer func() {
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
						w.server.sendWorkerError(err)
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

	w.stopRouters()
}

func (w *worker) stopRouters() {
	for _, router := range w.routers {
		router.stop()
	}
}
