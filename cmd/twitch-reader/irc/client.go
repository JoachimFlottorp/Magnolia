package irc

import (
	"strings"
	"sync"
	"time"

	"github.com/JoachimFlottorp/magnolia/cmd/twitch-reader/irc/parser"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type IrcConnection struct {
	Address  string
	User     string
	Password string

	Read 		chan string
	RecvPong 	chan bool
	MsgHasRecv 	chan bool

	Conn *websocket.Conn
	Mtx  sync.Mutex

	MessageSubscriber func(*parser.PrivmsgMessage)

	ConnectedChannels []string
}

func NewClient(username, password string) *IrcConnection {
	c := &IrcConnection{
		Address: CONNECTION_ADDRESS,
		User:     username,
		Password: password,

		Read: 		make(chan string),
		MsgHasRecv: make(chan bool),
		RecvPong: 	make(chan bool),

		Mtx: sync.Mutex{},

		ConnectedChannels: make([]string, 0),
	}

	return c
}

func (c *IrcConnection) OnMessage(cb func(msg *parser.PrivmsgMessage)) {
	c.MessageSubscriber = cb
}

func (c *IrcConnection) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.Address, nil)
	if err != nil { return err }

	c.Conn = conn
	wg := sync.WaitGroup{}

	// wg.Add(1)
	go c.handlePong(&wg)

	// wg.Add(1)
	go c.readLoop(&wg)

	go func() {
		for {
			msg := <-c.Read
			c.handleLine(msg)
		}
	}()

	c.Send("PASS " + c.Password)
	c.Send("NICK " + c.User)

	// wg.Wait()

	return nil
}

func (c *IrcConnection) Reconnect() {
	c.Conn.Close()

	c.Connect()
}

func (c *IrcConnection) handlePong(wg *sync.WaitGroup) {
	// defer func() {
	// 	wg.Done()
	// }()

	for {
		select {
		case <-c.MsgHasRecv: continue
		case <-time.After(4 * time.Minute): {
			zap.S().Debug("Sending PING")
			
			c.Send("PING : HI-:D")

			select {
			case <-c.RecvPong: continue
			case <-time.After(10 * time.Second): {
				zap.S().Errorw("Failed to receive pong from server")
			}
			}
		}
		}
	}
}

func (c *IrcConnection) readLoop(wg *sync.WaitGroup) {
	defer func() {
		c.Conn.Close()

		wg.Done()
	}()
	
	for {
		msgType, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.S().Errorw("Unexpected close from websocket error", "error", err)
			} else {
				zap.S().Errorw("Failed to read message from server", "error", err)
			}
			continue
		}

		if msgType != websocket.TextMessage {
			zap.S().Errorw("Received non-text message from server", "type", msgType)
			continue
		}

		lines := strings.Split(string(msg), "\r\n")
		for _, line := range lines {
			c.Read <- line
		}
	}
}

func (c *IrcConnection) handleLine(line string) {
	go func() {
		select {
		case c.MsgHasRecv <- true:
		default:
		}
	}()
	
	parsed, err := parser.ParseLine(line)
	if err != nil {
		zap.S().Errorw("Failed to parse line", "error", err)
		return
	}

	switch parsed.GetType() {
	case parser.PONG: {
		select {
		case c.RecvPong <- true:
		default:
		}
	}
	case parser.RECONNECT: {
		zap.S().Infow("Twitch told us to reconnect")
		c.Reconnect()
	}
	case parser.PING: {
		c.Send("PONG : HI-:D")
	}
	case parser.PRIVMSG: {
		msg := parsed.(*parser.PrivmsgMessage)
		if c.MessageSubscriber == nil { return }

		c.MessageSubscriber(msg)
	}
	}

}

func (c *IrcConnection) Join(channel string) {
	if channel == "" { return }
	
	c.Send("JOIN #" + channel)

	c.ConnectedChannels = append(c.ConnectedChannels, channel)
}

func (c *IrcConnection) Send(msg string) {
	c.Mtx.Lock()
	defer c.Mtx.Unlock()

	zap.S().Debugw("Sending message", "message", msg)
	
	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		zap.S().Errorw("Failed to send message to server", "error", err)
	}
}