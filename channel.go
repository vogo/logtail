package logtail

const DefaultChannelBufferSize = 16

type Message struct {
	ServerID string
	Data     []byte
}

type Channel chan *Message
