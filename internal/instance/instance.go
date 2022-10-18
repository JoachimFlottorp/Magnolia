package instance

import (
	"github.com/JoachimFlottorp/magnolia/internal/mongo"
	"github.com/JoachimFlottorp/magnolia/internal/redis"
)

type InstanceList struct {
	Redis 	redis.Instance
	Mongo 	mongo.Instance
}