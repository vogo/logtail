package logtail

import "bytes"

const TransferTypeDing = "ding"

type DingTransfer struct {
	url string
}

const dingMessageDataFixedBytesNum = 4

func (d *DingTransfer) Trans(serverID string, data ...[]byte) error {
	size := dingMessageDataFixedBytesNum + len(data)
	list := make([][]byte, size)
	list[0] = dingTextMessageDataPrefix
	list[1] = []byte(serverID)
	list[2] = messageTitleContentSplit

	for i, d := range data {
		list[i+3] = bytes.Replace(d, quotationBytes, escapeQuotationBytes, -1)
	}

	list[size-1] = dingTextMessageDataSuffix

	return httpTrans(d.url, list...)
}

func NewDingTransfer(url string) Transfer {
	return &DingTransfer{url: url}
}
