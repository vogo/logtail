package main

import (
	"flag"
	"os"
	"time"

	"github.com/vogo/logtail/repeater"
)

const readLogInterval = time.Millisecond * 10

func main() {
	var f = flag.String("f", "", "file")

	flag.Parse()

	if *f == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	c := make(chan []byte)
	go repeater.Repeat(*f, c)

	ticker := time.NewTicker(readLogInterval)

	for {
		b := <-c
		_, _ = os.Stdout.Write(b)

		<-ticker.C
	}
}
