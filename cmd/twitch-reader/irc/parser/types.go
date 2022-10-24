package parser

type MessageType int

const (
	UNSURE    MessageType = iota
	PRIVMSG   MessageType = iota
	PING      MessageType = iota
	PONG      MessageType = iota
	RECONNECT MessageType = iota
)

type Message interface {
	GetType() MessageType
}

type RawMessage struct {
	Raw  string
	Type MessageType
}

func (m *RawMessage) GetType() MessageType {
	return m.Type
}

type PrivmsgMessage struct {
	Raw     string
	Channel string
	User    string
	Message string
}

func (m *PrivmsgMessage) GetType() MessageType {
	return PRIVMSG
}

type PingMessage struct {
	Raw     string
	Message string
}

func (m *PingMessage) GetType() MessageType {
	return PING
}

type PongMessage struct {
	Raw     string
	Message string
}

func (m *PongMessage) GetType() MessageType {
	return PONG
}

type ReconnectMessage struct {
	Raw string
}

func (m *ReconnectMessage) GetType() MessageType {
	return RECONNECT
}