package logtail

import (
	"flag"
	"fmt"
	"os"

	"github.com/vogo/vogo/vos"
)

func Start() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			flag.PrintDefaults()
		}
	}()

	flag.Parse()
	config, err := parseConfig()
	if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
		os.Exit(1)
	}
	vos.LoadUserEnv()
	startLogtail(config)
}
