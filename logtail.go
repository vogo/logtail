package logtail

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/vogo/vogo/vio/vioutil"
)

var (
	port     = flag.Int("port", 54321, "tail port")
	path     = flag.String("path", "", "tail port")
	logFiles []string

	defaultLogtailWriter = &logtailWriter{}
)

func Start() {
	flag.Parse()
	if *port == 0 {
		*port = 54321
	}

	if *path == "" {
		panic("usage: logtail -port=<port> -path=<log_file_path>;<log_file_path>;<log_file_path>")
	}
	logFiles = strings.Split(*path, ";")
	for _, logFile := range logFiles {
		if !vioutil.ExistFile(logFile) {
			panic(fmt.Sprintf("%s not exists", logFile))
		}
	}

	startTailFiles(logFiles)

	handler := &httpHandler{
		writer: defaultLogtailWriter,
	}

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), handler); err != nil {
		panic(err)
	}
}
