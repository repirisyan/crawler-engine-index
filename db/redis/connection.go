package redisdb

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

func InitRedis() {
	Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),     // e.g. "localhost:6379"
		Password: os.Getenv("REDIS_PASSWORD"), // empty means no password
		DB:       0,                           // default DB
	})

	if _, err := Client.Ping(Ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("✅ Redis connected")
}

func GetRedisClient() *redis.Client {
	return Client
}
