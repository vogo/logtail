package logtail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vogo/logger"
)

type DingText struct {
	Content string `json:"content"`
}
type DingMessage struct {
	MsgType string   `json:"msgtype"`
	Text    DingText `json:"text"`
}

type DingAlerter struct {
	url string
}

func (d *DingAlerter) Alert(data []byte) error {
	msg := &DingMessage{
		MsgType: "text",
		Text: DingText{
			Content: fmt.Sprintf("logtail: %s", data),
		},
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	res, err := http.Post(d.url, "application/json", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			logger.Warnf("ding alert error: %s", body)
		}
		return fmt.Errorf("ding alert error, status code %d", res.StatusCode)
	}
	return nil
}

func NewDingAlerter(url string) Alerter {
	return &DingAlerter{url: url}
}
