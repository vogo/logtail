package logtail

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/vogo/logger"
)

func startTailFiles(files []string) {
	for _, filePath := range files {
		cmdStr := fmt.Sprintf("tail -f %s", filePath)

		go func() {
			for {
				logger.Info(cmdStr)
				cmd := exec.Command("/bin/sh", "-c", cmdStr)
				cmd.Stdout = defaultLogtailWriter
				cmd.Stderr = defaultLogtailWriter
				if err := cmd.Run(); err != nil {
					logger.Errorf("failed to tail file, try after 10s! error: %+v", err)
					time.Sleep(10 * time.Second)
				}
			}
		}()
	}

}
