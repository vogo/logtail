package logtail

import (
	"bytes"
	"sync/atomic"
	"time"
)

const TransferTypeDing = "ding"

type DingTransfer struct {
	url          string
	transferring int32 // whether transferring message
}

const dingMessageDataFixedBytesNum = 4
const dingMessageDataMaxLength = 1024

// transfer next message after the interval, ignore messages in the interval.
const dingMessageTransferInterval = time.Second * 5

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

		b = bytes.Replace(b, quotationBytes, escapeQuotationBytes, -1)
		b = bytes.Replace(b, []byte{'\n'}, []byte{'\\', 'n'}, -1)

		if len(b) > messageRemainCapacity {
			b = b[:messageRemainCapacity]
			for len(b) > 0 && b[len(b)-1]&0xC0 == 0x80 {
				b = b[:len(b)-1]
			}
		}

		list[idx] = b
		idx++

		messageRemainCapacity -= len(b)
	}

	list[idx] = dingTextMessageDataSuffix

	return httpTrans(d.url, list[:idx+1]...)
}

func NewDingTransfer(url string) Transfer {
	return &DingTransfer{
		url:          url,
		transferring: 0,
	}
}
