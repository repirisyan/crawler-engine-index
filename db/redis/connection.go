package redisdb

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

func InitRedis() {
	// InitRedis can run before InitPostgres (which also loads .env), so load it
	// here too - otherwise REDIS_HOST/REDIS_PORT are read before .env exists in
	// the process env and Addr silently becomes ":" (dial tcp :0).
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found (optional if using real env vars): %v", err)
	}

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	addr := fmt.Sprintf("%s:%s", host, port)

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"), // empty means no password
		DB:       0,                           // default DB
	})

	if _, err := Client.Ping(Ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", addr, err)
	}

	log.Printf("✅ Redis connected (%s)", addr)
}

func GetRedisClient() *redis.Client {
	return Client
}

func FlushAllDB() {
	// Flush all keys in all databases
	err := Client.FlushAll(Ctx).Err()
	if err != nil {
		log.Printf("⚠️ Redis flush failed: %v", err)
		return
	}

	log.Println("Redis flushed successfully")
}
