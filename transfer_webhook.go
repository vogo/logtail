package logtail

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vogo/logger"
	"github.com/vogo/vogo/vio"
)

const TransferTypeWebhook = "webhook"

var ErrHTTPStatusNonOK = errors.New("http status non ok")

type WebhookTransfer struct {
	url string
}

func (d *WebhookTransfer) Trans(_ string, data ...[]byte) error {
	return httpTrans(d.url, data...)
}

func (d *WebhookTransfer) start(*Router) error { return nil }

func NewWebhookTransfer(url string) Transfer {
	return &WebhookTransfer{url: url}
}

func httpTrans(url string, data ...[]byte) error {
	res, err := http.Post(url, "application/json", vio.NewBytesReader(data...))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			logger.Warnf("http alert error: %s", body)
		}

		return fmt.Errorf("http alert error, %w: %d", ErrHTTPStatusNonOK, res.StatusCode)
	}

	return nil
}
