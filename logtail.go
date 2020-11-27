package logtail

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/vogo/vogo/vos"
)

var (
	port    = flag.Int("port", 54321, "tail port")
	command = flag.String("cmd", "", "tail command")

	defaultLogtailWriter = &logtailWriter{
		lock:    sync.Mutex{},
		writers: make(map[int64]*websocketTransfer, 16),
	}
)

func Start() {
	flag.Parse()
	vos.LoadUserEnv()

	if *command == "" {
		fmt.Println("usage: logtail -port=<port> -cmd=<cmd>")
		os.Exit(1)
	}

	startTailCommand(*command)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), &httpHandler{}); err != nil {
		panic(err)
	}
}
