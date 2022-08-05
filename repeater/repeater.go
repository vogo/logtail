/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repeater

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

const filePerm = 0o600

func Repeat(filePath string, bytesChan chan []byte) {
	var (
		previousLine []byte
		line         []byte
		err          error
	)

	//nolint:nosnakecase // ignore snake case.
	file, err := os.OpenFile(filePath, os.O_RDONLY, filePerm)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed open file: %v", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(file)

	splitLine := []byte(`----------------`)

	for {
		var dataBuf []byte

		previousLine = nil

		for {
			line = readLine(reader)

			if len(previousLine) == 0 && bytes.Equal(line, splitLine) {
				dataBuf = dataBuf[:len(dataBuf)-1]

				break
			}

			dataBuf = append(dataBuf, line...)
			dataBuf = append(dataBuf, '\n')

			previousLine = line
		}

		bytesChan <- dataBuf
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
