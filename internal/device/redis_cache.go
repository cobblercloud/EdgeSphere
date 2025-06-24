package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	
	"github.com/go-redis/redis/v8"
	"edgesphere/internal/pkg/types"
)

type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	
	return &RedisCache{
		client: client,
		prefix: "edge:device:",
	}
}

func (c *RedisCache) key(id string) string {
	return c.prefix + id
}

// 使用BloomFilter检查存在性
func (c *RedisCache) Exists(id string) bool {
	ctx := context.Background()
	// 实际使用应替换为RedisBloom模块的BF.EXISTS
	return c.client.SIsMember(ctx, "device:bloom", id).Val()
}

func (c *RedisCache) BatchExists(ids []string) bool {
	ctx := context.Background()
	for _, id := range ids {
		if !c.client.SIsMember(ctx, "device:bloom", id).Val() {
			return false
		}
	}
	return true
}

func (c *RedisCache) Add(id string) {
	ctx := context.Background()
	c.client.SAdd(ctx, "device:bloom", id)
}

func (c *RedisCache) SetStatus(id string, status types.DeviceStatus) {
	ctx := context.Background()
	statusKey := c.key(id) + ":status"
	c.client.Set(ctx, statusKey, int(status), 10*time.Minute)
}

func (c *RedisCache) GetStatus(id string) types.DeviceStatus {
	ctx := context.Background()
	statusKey := c.key(id) + ":status"
	val, err := c.client.Get(ctx, statusKey).Int()
	if err != nil {
		return types.Unregistered
	}
	return types.DeviceStatus(val)
}

// 使用Redis Streams实现状态变更通知
func (c *RedisCache) PublishStatusUpdate(id string, status types.DeviceStatus) {
	ctx := context.Background()
	update := map[string]interface{}{
		"device_id": id,
		"status":    status,
		"timestamp": time.Now().Unix(),
	}
	c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: "device_status_updates",
		Values: update,
	})
}

// 订阅状态更新
func (c *RedisCache) SubscribeStatusUpdates() <-chan *redis.Message {
	ctx := context.Background()
	pubsub := c.client.Subscribe(ctx, "device_status_updates")
	return pubsub.Channel()
}