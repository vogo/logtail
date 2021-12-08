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

package main

import (
	"flag"
	"os"
	"time"

	"github.com/vogo/logtail/repeater"
)

const readLogInterval = time.Millisecond * 10

func main() {
	filePath := flag.String("f", "", "file")

	flag.Parse()

	if *filePath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	bytesChan := make(chan []byte)
	go repeater.Repeat(*filePath, bytesChan)

	ticker := time.NewTicker(readLogInterval)

	for {
		b := <-bytesChan
		_, _ = os.Stdout.Write(b)

		<-ticker.C
	}
}
