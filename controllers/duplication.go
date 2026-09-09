package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	redisdb "crawler-index/db/redis"
	Crawler "crawler-index/models/postgres/crawler"

	"github.com/joho/godotenv"
)

// RemoveDuplicationFromRedis drains the Redis set "products:crawler" into
// PostgreSQL in batches. Deduplication is handled by the
// `ON CONFLICT (title, seller_name, marketplace_id) DO NOTHING` constraint on
// the crawlers table, so there is no need to build an in-memory unique map of
// the entire set (which does not scale with crawl size).
func RemoveDuplicationFromRedis() {
	fmt.Println("🔍 Draining Redis set products:crawler into PostgreSQL")

	// Load environment variables (optional if real env vars are already set)
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ No .env file found: %v", err)
	}

	redisClient := redisdb.Client
	ctx := context.Background()

	const batchSize = 1000
	var cursor uint64
	batch := make([]Crawler.Product, 0, batchSize)
	saved := 0

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := Crawler.SaveProductsToPostgres(batch); err != nil {
			log.Printf("⚠️ Failed to save batch of %d products: %v", len(batch), err)
		} else {
			saved += len(batch)
		}
		batch = batch[:0]
	}

	for {
		values, newCursor, err := redisClient.SScan(ctx, "products:crawler", cursor, "*", int64(batchSize)).Result()
		if err != nil {
			log.Printf("❌ SSCAN failed, aborting: %v", err)
			break
		}

		for _, item := range values {
			var product Crawler.Product
			if err := json.Unmarshal([]byte(item), &product); err != nil {
				log.Printf("⚠️ Failed to parse product: %v", err)
				continue
			}

			if product.Weight == nil {
				zero := float32(0)
				product.Weight = &zero
			}

			batch = append(batch, product)
			if len(batch) >= batchSize {
				flush()
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	flush()

	fmt.Printf("✅ Saved %d products to PostgreSQL (duplicates skipped by DB constraint)\n", saved)
}
