package logtail

import (
	"flag"
	"fmt"
	"os"

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
	startLogtail(config)
}
