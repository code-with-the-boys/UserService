package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func Init() {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		Username:     "",
		MaxRetries:   5,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	RedisClient = redisClient
}

func GetRedisClient() *redis.Client {
	return RedisClient
}
