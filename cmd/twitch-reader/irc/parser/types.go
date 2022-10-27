package parser

type MessageType int

const (
	UNSURE    MessageType = iota
	PRIVMSG   MessageType = iota
	PING      MessageType = iota
	PONG      MessageType = iota
	RECONNECT MessageType = iota
	NOTICE    MessageType = iota
	ENDOFMOTD MessageType = iota
	JOIN      MessageType = iota
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

type NoticeMessage struct {
	Raw     string
	Channel string
	Message string
}

func (m *NoticeMessage) GetType() MessageType {
	return NOTICE
}

type EndOfMotdMessage struct {
	Raw     string
	User    string
	Message string
}

func (m *EndOfMotdMessage) GetType() MessageType {
	return ENDOFMOTD
}

type JoinMessage struct {
	Raw     string
	User    string
	Channel string
}

func (m *JoinMessage) GetType() MessageType {
	return JOIN
}
