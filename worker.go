package logtail

import (
	"fmt"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
)

type worker struct {
	id            string
	server        *Server
	sendErrorFlag bool
	command       string
	cmd           *exec.Cmd
	routerCount   int64
	routers       map[int64]*Router
}

func (w *worker) Write(bytes []byte) (int, error) {
	for _, r := range w.routers {
		r.receive(bytes)
	}

	_, _ = w.server.Write(bytes)

	return len(bytes), nil
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
		for {
			select {
			case <-w.server.stop:
				return
			default:
				logger.Infof("worker [%s] command: %s", w.id, w.command)
				w.cmd = exec.Command("/bin/sh", "-c", w.command)

				setCmdSysProcAttr(w.cmd)

				w.cmd.Stdout = w
				w.cmd.Stderr = w

				if err := w.cmd.Run(); err != nil {
					if w.sendErrorFlag {
						w.server.workerChan <- err
						return
					}

					select {
					case <-w.server.stop:
						return
					default:
						logger.Errorf("worker [%s] failed to exec command, retry after 10s! error: %+v, command: %s", w.id, err, w.command)
						time.Sleep(CommandFailRetryInterval)
					}
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
