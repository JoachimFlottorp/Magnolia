package instance

import (
	"github.com/JoachimFlottorp/yeahapi/internal/mongo"
	"github.com/JoachimFlottorp/yeahapi/internal/redis"
)

type InstanceList struct {
	Redis 	redis.Instance
	Mongo 	mongo.Instance
}