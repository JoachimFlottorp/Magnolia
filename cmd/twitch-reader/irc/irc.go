package irc

import (
	"sync"

	"github.com/JoachimFlottorp/magnolia/cmd/twitch-reader/irc/parser"
	"github.com/JoachimFlottorp/magnolia/internal/ctx"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

const (
	CONNECTION_ADDRESS = "wss://irc-ws.chat.twitch.tv:443"
	DEFAULT_USERNAME   = "justinfan123"
	DEFAULT_PASSWORD   = "XDDDDDD"

	MAX_CHANNELS_PER_CONN = 100
)

type IrcManager struct {
	ctx ctx.Context

	conns []*IrcConnection
	mtx   sync.Mutex

	joinQueue    chan string
	MessageQueue chan *parser.PrivmsgMessage
}

func NewManager(gCtx ctx.Context) *IrcManager {
	m := &IrcManager{
		ctx:          gCtx,
		conns:        make([]*IrcConnection, 0),
		mtx:          sync.Mutex{},
		joinQueue:    make(chan string),
		MessageQueue: make(chan *parser.PrivmsgMessage),
	}

	go func() {
		for {
			select {
			case <-gCtx.Done():
				return
			case channel := <-m.joinQueue:
				{
					conn, err := m.availableConnector()
					if err != nil {
						zap.S().Errorw("Failed to join channel", "channel", channel, "error", err)
						continue
					}

					conn.Join(channel)
				}
			}
		}
	}()

	return m
}

func (m *IrcManager) ConnectAllFromDatabase() error {
	cursor, err := m.ctx.Inst().Mongo.Collection(mongo.CollectionTwitch).Find(m.ctx, bson.D{})
	if err != nil {
		return err
	}

	channels := []mongo.TwitchChannel{}
	if err := cursor.All(m.ctx, &channels); err != nil {
		return err
	}

	for _, channel := range channels {
		conn, err := m.availableConnector()
		if err != nil {
			return err
		}

		conn.Join(channel.TwitchName)
	}

	return nil
}

func (m *IrcManager) availableConnector() (*IrcConnection, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, conn := range m.conns {
		if len(conn.ConnectedChannels) < MAX_CHANNELS_PER_CONN {
			return conn, nil
		}
	}

	return m.createNewConnector()
}

func (m *IrcManager) createNewConnector() (*IrcConnection, error) {
	conn := NewClient(DEFAULT_USERNAME, DEFAULT_PASSWORD)

	conn.OnMessage(func(msg *parser.PrivmsgMessage) {
		m.MessageQueue <- msg
	})

	m.conns = append(m.conns, conn)

	return conn, conn.Connect()
}

func (m *IrcManager) JoinChannel(channel mongo.TwitchChannel) {
	m.joinQueue <- channel.TwitchName
}
