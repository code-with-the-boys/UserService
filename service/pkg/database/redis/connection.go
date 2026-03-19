package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	ctx := context.Background()

	databaseURL := getEnv("REDIS_URL", "")

	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")
	dbnum := getEnv("REDIS_DB", "0")
	username := getEnv("REDIS_USERNAME", "")

	if databaseURL != "" {
		opt, err := redis.ParseURL(databaseURL)
		if err != nil {
			log.Fatalf("failed to parse REDIS_URL: %v", err)
		}
		RedisClient = redis.NewClient(opt)
		if err := RedisClient.Ping(ctx).Err(); err != nil {
			log.Fatalf("failed to connect to redis from REDIS_URL: %v", err)
		}
		return
	}

	db, err := strconv.Atoi(dbnum)
	if err != nil {
		log.Printf("invalid REDIS_DB value, fallback to 0: %v", err)
		db = 0
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     password,
		DB:           db,
		Username:     username,
		MaxRetries:   5,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	RedisClient = redisClient
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func GetRedisClient() *redis.Client {
	return RedisClient
}
