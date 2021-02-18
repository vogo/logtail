package logtail

import (
	"sync"

	"github.com/gorilla/websocket"
)

var (
	dingTextMessageDataPrefix = []byte(`{"msgtype":"text","text":{"content":"[logtail-`)
	dingTextMessageDataSuffix = []byte(`"}}`)
	messageTitleContentSplit  = []byte("]: ")
	quotationBytes            = []byte(`"`)
	escapeQuotationBytes      = []byte(`\"`)
)

var websocketUpgrader = websocket.Upgrader{}

var (
	serverDBLock = sync.Mutex{}
	serverDB     = make(map[string]*Server, 4)
)

var defaultFormat *Format
