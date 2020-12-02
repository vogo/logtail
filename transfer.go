package logtail

type transfer struct {
	index     int64
	transChan chan []byte
}
