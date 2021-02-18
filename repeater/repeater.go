package repeater

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

func Repeat(f string, c chan []byte) {
	var (
		previousLine []byte
		line         []byte
		err          error
	)

	file, err := os.OpenFile(f, os.O_RDONLY, 0o600)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed open file: %v", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(file)

	splitLine := []byte(`----------------`)

	for {
		var b []byte

		previousLine = nil

		for {
			line = readLine(reader)

			if len(previousLine) == 0 && bytes.Equal(line, splitLine) {
				b = b[:len(b)-1]

				break
			}

			b = append(b, line...)
			b = append(b, '\n')

			previousLine = line
		}

		c <- b
	}
}

func readLine(reader *bufio.Reader) []byte {
	var read []byte

	line, prefix, err := reader.ReadLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			// block when eof
			select {}
		}

		_, _ = fmt.Fprintf(os.Stderr, "\n\nread error: %v\n\n", err)
		os.Exit(1)
	}

	for prefix {
		read, prefix, err = reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// block when eof
				select {}
			}

			_, _ = fmt.Fprintf(os.Stderr, "\n\nread error: %v\n\n", err)
			os.Exit(1)
		}

		line = append(line, read...)
	}

	return line
}
