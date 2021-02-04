package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

type handler struct {
}

func (h *handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var (
		err  error
		data []byte
	)

	data, err = ioutil.ReadAll(req.Body)
	if err != nil {
		_, _ = res.Write([]byte(fmt.Sprintf("error: %v", err)))
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
