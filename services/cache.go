// Copyright 2024 OAuth Server Authors.
// Licensed under the Apache License, Version 2.0

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

// InitRedis initializes Redis connection
func InitRedis() error {
	redisAddr, _ := web.AppConfig.String("redisAddr")
	redisPassword, _ := web.AppConfig.String("redisPassword")
	redisDB, _ := web.AppConfig.Int("redisDB")

	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return nil
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
