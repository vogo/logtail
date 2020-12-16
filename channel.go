package logtail

type Message struct {
	ServerId string
	Data     []byte
}

type Channel chan *Message
