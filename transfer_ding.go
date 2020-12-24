package logtail

import (
	"encoding/json"
	"fmt"
)

const TransferTypeDing = "ding"

type DingText struct {
	Content string `json:"content"`
}
type DingMessage struct {
	MsgType string   `json:"msgtype"`
	Text    DingText `json:"text"`
}

type DingTransfer struct {
	url string
}

func (d *DingTransfer) Trans(serverID string, data []byte) error {
	msg := &DingMessage{
		MsgType: "text",
		Text: DingText{
			Content: fmt.Sprintf("[logtail-%s]: %s", serverID, data),
		},
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return httpTrans(d.url, jsonBytes)
}

func NewDingTransfer(url string) Transfer {
	return &DingTransfer{url: url}
}
