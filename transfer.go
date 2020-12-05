package logtail

type Transfer interface {
	Trans(data []byte) error
}
