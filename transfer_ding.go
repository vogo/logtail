package logtail

import (
	"sync/atomic"
	"time"

	"github.com/vogo/logger"
)

const TransferTypeDing = "ding"

const (
	dingMessageDataFixedBytesNum = 4
	dingMessageDataMaxLength     = 1024
)

// transfer next message after the interval, ignore messages in the interval.
const dingMessageTransferInterval = time.Second * 5

type DingTransfer struct {
	url          string
	transferring int32 // whether transferring message
}

func (d *DingTransfer) start(*Router) error { return nil }

func (d *DingTransfer) Trans(serverID string, data ...[]byte) error {
	if !atomic.CompareAndSwapInt32(&d.transferring, 0, 1) {
		// ignore message to
		return nil
	}

	go func() {
		<-time.After(dingMessageTransferInterval)
		atomic.StoreInt32(&d.transferring, 0)
	}()

	size := dingMessageDataFixedBytesNum + len(data)
	list := make([][]byte, size)
	list[0] = dingTextMessageDataPrefix
	list[1] = []byte(serverID)
	list[2] = messageTitleContentSplit

	idx := 3
	messageRemainCapacity := dingMessageDataMaxLength

	for _, b := range data {
		if messageRemainCapacity <= 0 {
			break
		}

		b = EscapeLimitJSONBytes(b, messageRemainCapacity)

		list[idx] = b
		idx++

		messageRemainCapacity -= len(b)
	}

	list[idx] = dingTextMessageDataSuffix

	if err := httpTrans(d.url, list[:idx+1]...); err != nil {
		logger.Errorf("ding error: %v", err)
	}

	return nil
}

func NewDingTransfer(url string) Transfer {
	return &DingTransfer{
		url:          url,
		transferring: 0,
	}
}
