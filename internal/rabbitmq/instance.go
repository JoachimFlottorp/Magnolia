package rabbitmq

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type ExchangeType string
const (
	ExchangeTypeDirect 	= ExchangeType("direct")
	ExchangeTypeFanout 	= ExchangeType("fanout")
	ExchangeTypeTopic  	= ExchangeType("topic")
	ExchangeTypeHeaders = ExchangeType("headers")
)


type PublishSettings struct {
	Exchange    string
	RoutingKey  string
	Msg 	    amqp091.Publishing
}

type ConsumeSettings struct {
	Queue 		string
	Consumer	string
}

type QueueSettings struct {
	Name       string
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
	// Consume messages from a queue
	//
	// Responds with a channel that sends the result of the deserialization  
	Consume(context.Context, ConsumeSettings) (chan amqp091.Delivery, error)
}