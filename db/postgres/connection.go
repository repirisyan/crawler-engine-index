package pgdb

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	PostgresPool *pgxpool.Pool
	Ctx          = context.Background()
)

func InitPostgres() {
	err := godotenv.Load() // load .env dari root project
	if err != nil {
		log.Println("No .env file found (optional if using real env vars)")
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"), // host:port
		os.Getenv("POSTGRES_DB"),
	)

	// Parse the DSN to get the configuration
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Failed to parse PostgreSQL DSN: %v", err)
	}

	// Create the connection pool
	PostgresPool, err = pgxpool.NewWithConfig(Ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v", err)
	}

	// Test the connection
	err = PostgresPool.Ping(Ctx)
	if err != nil {
		log.Fatalf("PostgreSQL ping failed: %v", err)
	}

	log.Println("✅ PostgreSQL connected")
}

func GetPostgresPool() *pgxpool.Pool {
	return PostgresPool
}
