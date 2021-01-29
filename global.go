package logtail

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var dingTextMessageDataPrefix = []byte(`{"msgtype":"text","text":{"content":"[logtail-`)
var dingTextMessageDataSuffix = []byte(`"}}`)
var messageTitleContentSplit = []byte("]: ")
var quotationBytes = []byte(`"`)
var escapeQuotationBytes = []byte(`\"`)

var websocketUpgrader = websocket.Upgrader{}

var serverDBLock = sync.Mutex{}
var serverDB = make(map[string]*Server, 4)

var defaultFormat *Format

var routerReadTicker = time.NewTicker(time.Millisecond * 50)
