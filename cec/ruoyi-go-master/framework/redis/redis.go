package redis

import (
	"context"
	"time"

	"backend/framework/dal"

	"github.com/go-redis/redis/v8"
)

// GetClient 获取 Redis 客户端实例
func GetClient() *redis.Client {
	return dal.Redis
}

// Set 设置键值对
func Set(key string, value interface{}, expiration time.Duration) error {
	return GetClient().Set(context.Background(), key, value, expiration).Err()
}

// Get 获取键值
func Get(key string) (string, error) {
	return GetClient().Get(context.Background(), key).Result()
}

// Del 删除键
func Del(key string) error {
	return GetClient().Del(context.Background(), key).Err()
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	result, err := GetClient().Exists(context.Background(), key).Result()
	return result > 0, err
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return GetClient().Expire(context.Background(), key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func TTL(key string) (time.Duration, error) {
	return GetClient().TTL(context.Background(), key).Result()
}
