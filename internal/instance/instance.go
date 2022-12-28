package instance

import (
	"context"

	"github.com/JoachimFlottorp/magnolia/internal/config"
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/rabbitmq"
	"github.com/JoachimFlottorp/magnolia/internal/redis"
)

type InstanceList struct {
	Redis redis.Instance
	Mongo mongo.Instance
	RMQ   rabbitmq.Instance
}

func CreateRedisInstance(ctx context.Context, conf *config.Config) (redis.Instance, error) {
	return redis.Create(ctx, redis.Options{
		Address:  conf.Redis.Address,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		DB:       conf.Redis.Database,
	})
}

func CreateMongoInstance(ctx context.Context, conf *config.Config) (mongo.Instance, error) {
	return mongo.New(ctx, conf)
}

func CreateRabbitMQInstance(ctx context.Context, conf *config.Config) (rabbitmq.Instance, error) {
	return rabbitmq.New(ctx, &rabbitmq.NewInstanceSettings{
		Address: conf.RabbitMQ.URI,
	})
}
