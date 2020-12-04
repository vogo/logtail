package logtail

type Alerter interface {
	Alert(data []byte) error
}
