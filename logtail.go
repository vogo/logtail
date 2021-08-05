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

package logtail

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vogo/logger"
	"github.com/vogo/vogo/vos"
)

// Start parse command config, and start logtail servers with http listener.
func Start() {
	config, err := parseConfig()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		flag.PrintDefaults()
		os.Exit(1)
	}

	configLogLevel(config.LogLevel)

	vos.LoadUserEnv()

	// stop exist servers first
	StopLogtail()

	go StartLogtail(config)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), &httpHandler{}); err != nil {
			panic(err)
		}
	}()

	handleSignal()
}

func configLogLevel(level string) {
	level = strings.ToUpper(level)
	switch level {
	case "ERROR":
		logger.SetLevel(logger.LevelError)
	case "WARN":
		logger.SetLevel(logger.LevelWarn)
	case "INFO":
		logger.SetLevel(logger.LevelInfo)
	case "DEBUG":
		logger.SetLevel(logger.LevelDebug)
	}
}

func handleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-signalChan
	logger.Infof("signal: %v", sig)
	StopLogtail()

	// wait all goroutines stopping
	<-time.After(time.Second)
}
