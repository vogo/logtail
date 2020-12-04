package logtail

import (
	"sync/atomic"
)

type Tunnel chan []byte

var tunnelCount int64 = 0

func AddTunnel(t Tunnel) int64 {
	index := atomic.AddInt64(&tunnelCount, 1)
	defaultLogtailWriter.addTunnel(index, t)
	return index
}

func RemoveTunnel(index int64) {
	defaultLogtailWriter.removeTunnel(index)
}
