package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Options struct {
	Address  string
	Username string
	Password string
	DB       int
}

type Instance interface {
	// Ping checks if the redis instance is alive
	Ping(context.Context) error

	// Get returns the value of the key
	Get(context.Context, string) (string, error)
	// Set sets the value of the key
	//
	// No expiration is set
	Set(context.Context, string, string) error
	// Del deletes the key
	Del(context.Context, string) error
	// Expire sets the expiration of the key
	Expire(context.Context, string, time.Duration) error
	
	// Prefix returns the prefix used for all keys
	Prefix() string

	Client() *redis.Client
}
