package logtail

type Transfer interface {
	Trans(serverId string, data []byte) error
}
