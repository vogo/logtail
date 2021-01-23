package logtail

type Transfer interface {
	Trans(serverID string, data ...[]byte) error
}
