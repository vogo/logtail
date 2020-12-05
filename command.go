package logtail

import (
	"os/exec"
	"time"

	"github.com/vogo/logger"
)

func startTailCommand(cmdStr string) {
	go func() {
		for {
			logger.Info(cmdStr)
			cmd := exec.Command("/bin/sh", "-c", cmdStr)
			cmd.Stdout = defaultLogtailServer
			cmd.Stderr = defaultLogtailServer

			if err := cmd.Run(); err != nil {
				logger.Errorf("failed to tail file, try after 10s! error: %+v", err)
				time.Sleep(10 * time.Second)
			}
		}
	}()

}
