// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/oauth-server/oauth-server/config"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient     *redis.Client
	ctx             = context.Background()
	ServerStartTime = time.Now() // 服务器启动时间
)

const (
	// Cache expiration times
	UserCacheExpiration        = 15 * time.Minute // 用户信息缓存 15 分钟
	ApplicationCacheExpiration = 30 * time.Minute // 应用配置缓存 30 分钟
	TokenCacheExpiration       = 1 * time.Hour    // Token 缓存 1 小时
)

// InitRedis initializes Redis connection
func InitRedis() error {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	redisCfg := cfg.Redis
	redisAddr := fmt.Sprintf("%s:%s", redisCfg.Host, redisCfg.Port)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	// Test connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return nil
}

// PingRedis 检查 Redis 连接状态
func PingRedis() error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return redisClient.Ping(ctx).Err()
}

// GetRedisClient returns the Redis client
func GetRedisClient() *redis.Client {
	return redisClient
}

// CacheToken caches a token in Redis
func CacheToken(tokenHash string, tokenData interface{}, expiration time.Duration) error {
	if redisClient == nil {
		return nil // Redis not configured
	}

	// Use default expiration if not specified
	if expiration == 0 {
		expiration = TokenCacheExpiration
	}

	data, err := json.Marshal(tokenData)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, fmt.Sprintf("token:%s", tokenHash), data, expiration).Err()
}

// GetCachedToken retrieves a cached token from Redis
func GetCachedToken(tokenHash string) ([]byte, error) {
	if redisClient == nil {
		return nil, nil // Redis not configured
	}

	data, err := redisClient.Get(ctx, fmt.Sprintf("token:%s", tokenHash)).Bytes()
	if err == redis.Nil {
		return nil, nil // Key does not exist
	}
	return data, err
}

// DeleteCachedToken removes a token from cache
func DeleteCachedToken(tokenHash string) error {
	if redisClient == nil {
		return nil
	}

	return redisClient.Del(ctx, fmt.Sprintf("token:%s", tokenHash)).Err()
}

// CacheUser caches user data
func CacheUser(userId string, userData interface{}, expiration time.Duration) error {
	if redisClient == nil {
		return nil
	}

	// Use default expiration if not specified
	if expiration == 0 {
		expiration = UserCacheExpiration
	}

	data, err := json.Marshal(userData)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, fmt.Sprintf("user:%s", userId), data, expiration).Err()
}

// GetCachedUser retrieves cached user data
func GetCachedUser(userId string) ([]byte, error) {
	if redisClient == nil {
		return nil, nil
	}

	data, err := redisClient.Get(ctx, fmt.Sprintf("user:%s", userId)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// InvalidateUserCache removes user from cache
func InvalidateUserCache(userId string) error {
	if redisClient == nil {
		return nil
	}

	return redisClient.Del(ctx, fmt.Sprintf("user:%s", userId)).Err()
}

// CacheApplication caches application configuration
func CacheApplication(clientId string, appData interface{}) error {
	if redisClient == nil {
		return nil
	}

	data, err := json.Marshal(appData)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, fmt.Sprintf("app:%s", clientId), data, ApplicationCacheExpiration).Err()
}

// GetCachedApplication retrieves cached application configuration
func GetCachedApplication(clientId string) ([]byte, error) {
	if redisClient == nil {
		return nil, nil
	}

	data, err := redisClient.Get(ctx, fmt.Sprintf("app:%s", clientId)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return data, err
}

// InvalidateApplicationCache removes application from cache
func InvalidateApplicationCache(clientId string) error {
	if redisClient == nil {
		return nil
	}

	return redisClient.Del(ctx, fmt.Sprintf("app:%s", clientId)).Err()
}

// SetRateLimit sets rate limit counter
func SetRateLimit(key string, limit int, window time.Duration) error {
	if redisClient == nil {
		return nil
	}

	pipe := redisClient.Pipeline()
	pipe.Incr(ctx, fmt.Sprintf("ratelimit:%s", key))
	pipe.Expire(ctx, fmt.Sprintf("ratelimit:%s", key), window)
	_, err := pipe.Exec(ctx)
	return err
}

// CheckRateLimit checks if rate limit is exceeded
func CheckRateLimit(key string, limit int) (bool, error) {
	if redisClient == nil {
		return false, nil // No rate limiting if Redis not configured
	}

	count, err := redisClient.Get(ctx, fmt.Sprintf("ratelimit:%s", key)).Int()
	if err == redis.Nil {
		return false, nil // No limit set yet
	}
	if err != nil {
		return false, err
	}

	return count >= limit, nil
}

// ClearCache clears all cache entries in Redis
func ClearCache() error {
	if redisClient == nil {
		return nil // Redis not configured
	}

	return redisClient.FlushDB(ctx).Err()
}
