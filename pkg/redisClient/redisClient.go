// Package redisClient -----------------------------
// @file      : redisClient.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/2/24 19:03
// -------------------------------------------
package redisClient

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisClient 结构体用于封装 Redis 客户端
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 用于创建一个新的 Redis 客户端
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:           addr,
		Password:       password,
		DB:             db,
		ReadTimeout:    -2,
		MaxActiveConns: 50,
		MinIdleConns:   5,
		MaxIdleConns:   10,
	})

	// 检查连接是否正常
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		fmt.Printf("failed to connect to Redis: %v", err)
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}
func (r *RedisClient) Client() *redis.Client {
	return r.client
}
func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) HMSet(ctx context.Context, key string, fields map[string]interface{}) error {
	if r == nil {
		return errors.New("redis client nill")
	}
	return r.client.HMSet(ctx, key, fields).Err()
}
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	if r == nil {
		return errors.New("redis client nil")
	}
	return r.client.HDel(ctx, key, fields...).Err()
}
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}
func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

func (r *RedisClient) SetWithTimeout(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *RedisClient) PopFromListR(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(ctx, key, members...).Result()
}

// SIsMember 检查成员是否存在于集合中
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) (*redis.PubSub, error) {
	pubsub := r.client.Subscribe(ctx, channels...)
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to channels: %v", err)
	}
	return pubsub, nil
}

func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	result := r.client.Publish(ctx, channel, message)
	if result.Err() != nil {
		return fmt.Errorf("failed to publish message: %v", result.Err())
	}
	return nil
}

func (r *RedisClient) AddToList(ctx context.Context, key string, values ...interface{}) (int64, error) {
	result, err := r.client.RPush(ctx, key, values...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to add values to list: %v", err)
	}
	return result, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	if r == nil {
		fmt.Println("redis r is nil")
		return errors.New("Redis client is not initialized")
	}
	_, err := r.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis Ping失败: %v", err)
	}
	return nil
}
