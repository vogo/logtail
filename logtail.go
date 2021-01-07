package logtail

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vogo/logger"
	"github.com/vogo/vogo/vos"
)

func Start() {
	config, err := parseConfig()
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	vos.LoadUserEnv()

	go startLogtail(config)

	handleSignal()
}

func handleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-signalChan
	logger.Infof("signal: %v", sig)
	stopServers()
}
