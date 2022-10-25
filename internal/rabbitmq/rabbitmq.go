package rabbitmq

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	ErrNotDeclared = errors.New("queue not declared")
)

type rabbitmqInstance struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	isOpen 	chan *amqp.Error
}

func New(ctx context.Context, opts *NewInstanceSettings) (Instance, error) {
	conn, err := amqp.Dial(opts.Address)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		ch.Close()
		conn.Close()
	}()

	return &rabbitmqInstance{
		conn:    conn,
		channel: ch,
		isOpen:  conn.NotifyClose(make(chan *amqp.Error)),
	}, nil
}

func (r *rabbitmqInstance) CreateQueue(ctx context.Context, opts QueueSettings) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		opts.Name.String(),
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *rabbitmqInstance) CreateExchange(ctx context.Context, opts ExchangeSettings) error {
	return r.channel.ExchangeDeclare(
		opts.Name,
		string(opts.Type),
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *rabbitmqInstance) BindQueue(ctx context.Context, opts BindingSettings) error {
	return r.channel.QueueBind(
		opts.Name,
		opts.RoutingKey,
		opts.Exchange,
		false,
		nil,
	)
}

func (r *rabbitmqInstance) Publish(ctx context.Context, opts PublishSettings) error {
	err := r.channel.PublishWithContext(
		ctx,
		opts.Exchange,
		opts.RoutingKey.String(),
		false,
		false,
		opts.Msg,
	)

	if err != nil {
		zap.S().Errorw("Failed to publish to RabbitMQ", "error", err)
	}

	return err
}

func (r *rabbitmqInstance) Consume(ctx context.Context, opts ConsumeSettings) (chan *amqp.Delivery, error) {
	_, err := r.CreateQueue(ctx, QueueSettings{
		Name: opts.Queue,
	})

	if err != nil { return nil, err }
	
	
	msgs, err := r.channel.Consume(
		opts.Queue.String(),
		opts.Consumer,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil { return nil, err }

	out := make(chan *amqp.Delivery, 50)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(out)
				return
			case <-r.isOpen:
				zap.S().Warn("RabbitMQ connection closed")
				close(out)
				return
			case msg, ok := <-msgs: {
				if !ok {
					zap.S().Errorw("Channel is not ok", "queue", opts.Queue)
					close(out)
					return
				}

				// TODO figure out a way to automatically 
				// serialize the message from json or protobuf to a struct
				// Automatically if this is even possible.

				out <- &msg

				msg.Ack(false)
			}
			}
		}
	}()

	return out, nil	
}