package logtail

const DefaultChannelBufferSize = 16

type Message struct {
	Server *Server
	Data   []byte
}

type Channel chan *Message
