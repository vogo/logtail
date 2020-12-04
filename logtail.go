package logtail

import (
	"flag"
	"fmt"
	"os"

	"github.com/vogo/vogo/vos"
)

var (
	port          = flag.Int("port", 54321, "tail port")
	command       = flag.String("cmd", "", "tail command")
	matchContains = flag.String("match_contains", "", "a containing string")
	alertUrl      = flag.String("alert_url", "", "alert ding url")
)

func Start() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			flag.PrintDefaults()
		}
	}()

	flag.Parse()
	vos.LoadUserEnv()

	if *command == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *matchContains != "" && *alertUrl != "" {
		filter := NewFilter(NewContainsMatcher(*matchContains), NewDingAlerter(*alertUrl))
		NewTransfer(filter).Start()
	}

	startTailCommand(*command)

	startHttpListener()
}
