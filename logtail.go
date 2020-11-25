package logtail

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/vogo/vogo/vos"
)

var (
	port    = flag.Int("port", 54321, "tail port")
	command = flag.String("cmd", "", "tail command")

	defaultLogtailWriter = &logtailWriter{}
)

func Start() {
	flag.Parse()
	vos.LoadUserEnv()

	if *port == 0 {
		*port = 54321
	}

	if *command == "" {
		panic("usage: logtail -port=<port> -cmd=<cmd>")
	}

	startTailCommand(*command)

	handler := &httpHandler{
		writer: defaultLogtailWriter,
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), handler); err != nil {
		panic(err)
	}
}
