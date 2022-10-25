package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type ExchangeType 	string
type QueueName 		string

func (q QueueName) String() string {
	return string(q)
}

const (
	ExchangeTypeDirect 	= ExchangeType("direct")
	ExchangeTypeFanout 	= ExchangeType("fanout")
	ExchangeTypeTopic  	= ExchangeType("topic")
	ExchangeTypeHeaders = ExchangeType("headers")

	QueueJoinRequest 		= QueueName("twitch-join-request")
	QueueMarkovGenenerator 	= QueueName("markov-generator")
)


type PublishSettings struct {
	Exchange    string
	RoutingKey  QueueName
	Msg 	    amqp091.Publishing
}

type ConsumeSettings struct {
	Queue 		QueueName
	Consumer	string
}

type QueueSettings struct {
	Name       QueueName
}

type ExchangeSettings struct {
	Name       string
	Type 	   ExchangeType
}

type BindingSettings struct {
	Name string
	RoutingKey string
	Exchange string
}

type NewInstanceSettings struct {
	Address string
}

type Instance interface {
	// Publish a message to the specified exchange
	Publish(context.Context, PublishSettings) error
	// Create a queue
	CreateQueue(context.Context, QueueSettings) (amqp091.Queue, error)
	CreateExchange(context.Context, ExchangeSettings) error
	BindQueue(context.Context, BindingSettings) error
	Consume(context.Context, ConsumeSettings) (chan *amqp091.Delivery, error)
}