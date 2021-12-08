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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DingText struct {
	Content string `json:"content"`
}

type DingMessage struct {
	MsgType string   `json:"msgtype"`
	Text    DingText `json:"text"`
}

type handler struct{}

func (h *handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var (
		err  error
		data []byte
	)

	data, err = io.ReadAll(req.Body)
	if err != nil {
		_, _ = fmt.Fprintf(res, "error: %v", err)

		return
	}

	msg := &DingMessage{}

	err = json.Unmarshal(data, msg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "json unmarshal error: %v, data: %s\n", err, data)

		return
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", msg.Text.Content)
	_, _ = res.Write([]byte("ok"))
}

func main() {
	if err := http.ListenAndServe(":55321", &handler{}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}
