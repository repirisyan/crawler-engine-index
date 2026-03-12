package redisdb

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

func InitRedis() {
	addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,                        // e.g. "localhost:6379"
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

func FlushAllDB() {
	// Flush all keys in all databases
	err := Client.FlushAll(Ctx).Err()
	if err != nil {
		panic(err)
	}

	log.Println("Redis flushed successfully")
}
