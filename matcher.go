package logtail

type Matcher interface {
	Match(bytes []byte) bool
}
