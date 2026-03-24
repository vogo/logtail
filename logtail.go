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
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vogo/logtail/internal/starter"
	"github.com/vogo/logtail/internal/tail"
	"github.com/vogo/logtail/internal/webapi"
	"github.com/vogo/vogo/vlog"
)

func main() {
	tailer, err := starter.Start()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}

	webapi.StartWebAPI(tailer)

	handleSignal(tailer)
}

func handleSignal(tailer *tail.Tailer) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-signalChan
	vlog.Infof("signal: %v", sig)

	tailer.Stop()

	// wait all goroutines stopping
	<-time.After(time.Second)
}
