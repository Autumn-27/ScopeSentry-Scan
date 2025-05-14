// redis-------------------------------------
// @file      : redis.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 21:47
// -------------------------------------------

package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/redis/go-redis/v9"
	"time"
)

// Client 结构体用于封装 Redis 客户端
type Client struct {
	client *redis.Client
}

var RedisClient *Client

//// NewRedisConnect 用于创建一个新的 Redis 客户端
//func NewRedisConnect(ip, port, password string, db int) (*Client, error) {
//	client := redis.NewClient(&redis.Options{
//		Addr:           ip + ":" + port,
//		Password:       password,
//		DB:             db,
//		ReadTimeout:    -2,
//		MaxActiveConns: 50,
//		MinIdleConns:   5,
//		MaxIdleConns:   10,
//	})
//
//	// 检查连接是否正常
//	_, err := client.Ping(context.Background()).Result()
//	if err != nil {
//		fmt.Printf("failed to connect to Redis: %v", err)
//		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
//	}
//
//	return &Client{client: client}, nil
//}

func Initialize() {
	client := redis.NewClient(&redis.Options{
		Addr:           global.AppConfig.Redis.IP + ":" + global.AppConfig.Redis.Port,
		Password:       global.AppConfig.Redis.Password,
		DB:             0,
		ReadTimeout:    -2,
		MaxActiveConns: 50,
		MinIdleConns:   5,
		MaxIdleConns:   10,
	})

	// 检查连接是否正常
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		fmt.Printf("failed to connect to Redis: %v", err)
	}

	RedisClient = &Client{client: client}
}

func (r *Client) Client() *redis.Client {
	return r.client
}

func (r *Client) Close() error {
	return r.client.Close()
}

func (r *Client) HMSet(ctx context.Context, key string, fields map[string]interface{}) error {
	if r == nil {
		return errors.New("redis client nill")
	}
	return r.client.HMSet(ctx, key, fields).Err()
}
func (r *Client) HDel(ctx context.Context, key string, fields ...string) error {
	if r == nil {
		return errors.New("redis client nil")
	}
	return r.client.HDel(ctx, key, fields...).Err()
}

func (r *Client) Del(ctx context.Context, key string) error {
	if r == nil {
		return errors.New("redis client nil")
	}
	return r.client.Del(ctx, key).Err()
}

func (r *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}
func (r *Client) HSet(ctx context.Context, key, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

func (r *Client) SetWithTimeout(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}
func (r *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *Client) PopFromListR(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

func (r *Client) GetFirstFromList(ctx context.Context, key string) (string, error) {
	return r.client.LIndex(ctx, key, 0).Result()
}

func (r *Client) PopFirstFromList(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

func (r *Client) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(ctx, key, members...).Result()
}

func (r *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// LLen 获取列表的长度
func (r *Client) LLen(ctx context.Context, key string) (int64, error) {
	if r.client == nil {
		return 0, errors.New("redis client nil")
	}
	return r.client.LLen(ctx, key).Result()
}

// LRange 获取列表中指定范围的元素
func (r *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if r.client == nil {
		return nil, errors.New("redis client nil")
	}
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LRem 删除列表中指定数量的元素
func (r *Client) LRem(ctx context.Context, key string, count int64, value string) error {
	if r.client == nil {
		return errors.New("redis client nil")
	}
	return r.client.LRem(ctx, key, count, value).Err()
}

// BatchGetAndDelete 批量获取并删除列表元素
func (r *Client) BatchGetAndDelete(ctx context.Context, key string, count int64) ([]string, error) {
	if r.client == nil {
		return nil, errors.New("redis client nil")
	}

	// 获取列表中指定数量的元素
	elements, err := r.LRange(ctx, key, 0, count-1)
	if err != nil {
		return nil, err
	}

	// 使用管道批量删除获取到的元素
	pipe := r.client.Pipeline()
	for _, element := range elements {
		pipe.LRem(ctx, key, 0, element) // 0 表示删除所有匹配的元素
	}
	_, err = pipe.Exec(ctx) // 执行管道
	if err != nil {
		return nil, err
	}

	return elements, nil
}

// SIsMember 检查成员是否存在于集合中
func (r *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}
func (r *Client) Subscribe(ctx context.Context, channels ...string) (*redis.PubSub, error) {
	pubsub := r.client.Subscribe(ctx, channels...)
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to channels: %v", err)
	}
	return pubsub, nil
}

func (r *Client) Publish(ctx context.Context, channel string, message interface{}) error {
	result := r.client.Publish(ctx, channel, message)
	if result.Err() != nil {
		return fmt.Errorf("failed to publish message: %v", result.Err())
	}
	return nil
}

func (r *Client) AddToList(ctx context.Context, key string, values ...interface{}) (int64, error) {
	result, err := r.client.RPush(ctx, key, values...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to add values to list: %v", err)
	}
	return result, nil
}

func (r *Client) Set(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *Client) Ping(ctx context.Context) error {
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
