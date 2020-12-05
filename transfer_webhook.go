package logtail

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vogo/logger"
)

type WebhookTransfer struct {
	url string
}

func (d *WebhookTransfer) Trans(data []byte) error {
	return httpTrans(d.url, data)
}

func NewWebhookTransfer(url string) Transfer {
	return &WebhookTransfer{url: url}
}

func httpTrans(url string, data []byte) error {
	res, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			logger.Warnf("http alert error: %s", body)
		}
		return fmt.Errorf("http alert error, status code %d", res.StatusCode)
	}
	return nil
}
