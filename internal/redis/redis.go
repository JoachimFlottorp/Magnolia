package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisInstance struct {
	client *redis.Client
}

func Create(ctx context.Context, options Options) (Instance, error) {
	rds := redis.NewClient(&redis.Options{
		Addr:     options.Address,
		Username: options.Username,
		Password: options.Password,
		DB:       options.DB,
	})

	if err := rds.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	inst := &redisInstance{
		client: rds,
	}

	return inst, nil
}

func (r *redisInstance) formatKey(key string) string {
	return fmt.Sprintf("%s%s", r.Prefix(), key)
}

func (r *redisInstance) Prefix() string {
	return "magnolia:"
}

func (r *redisInstance) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisInstance) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, r.formatKey(key)).Result()
}

func (r *redisInstance) Set(ctx context.Context, key string, value string) error {
	return r.client.Set(ctx, r.formatKey(key), value, 0).Err()
}

func (r *redisInstance) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.formatKey(key)).Err()
}

func (r *redisInstance) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, r.formatKey(key), expiration).Err()
}

func (r *redisInstance) LPush(ctx context.Context, key string, value string) error {
	return r.client.LPush(ctx, r.formatKey(key), value).Err()
}

func (r *redisInstance) LRPop(ctx context.Context, key string) error {
	return r.client.RPop(ctx, r.formatKey(key)).Err()
}

func (r *redisInstance) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, r.formatKey(key)).Result()
}

func (r *redisInstance) GetAllList(ctx context.Context, key string) ([]string, error) {
	a, e := r.client.LRange(ctx, r.formatKey(key), 0, -1).Result()

	if len(a) == 0 {
		return nil, redis.Nil
	}

	return a, e
}

func (r *redisInstance) Client() *redis.Client {
	return r.client
}
