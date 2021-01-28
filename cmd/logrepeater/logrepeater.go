package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

const readLogInterval = time.Millisecond * 10

func main() {
	var f = flag.String("f", "", "file")

	flag.Parse()

	if *f == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	file, err := os.OpenFile(*f, os.O_RDONLY, 0600)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed open file: %v", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(file)

	repeat(reader)
}

func repeat(reader *bufio.Reader) {
	var (
		b      []byte
		prefix bool
		line   []byte
		err    error
	)

	var splitLine = []byte(`----------------`)

	ticker := time.NewTicker(readLogInterval)

	for {
		b = b[:0]

		prefix = true

		for prefix {
			<-ticker.C

			line, prefix, err = reader.ReadLine()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "\n\nread error: %v\n\n", err)

				if err == io.EOF {
					// block until being killed
					select {}
				}

				return
			}

			if len(line) == 0 || bytes.Equal(line, splitLine) {
				prefix = true
				continue
			}

			if len(b) > 0 {
				b = append(b, '\n')
			}

			b = append(b, line...)
		}

		b = append(b, '\n')

		_, _ = os.Stdout.Write(b)
	}
}
