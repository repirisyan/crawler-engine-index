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

func RemoveDuplicationFromRedis() {
	fmt.Println("🔍 Removing duplicates from Redis set: products:crawler")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("❌ Error loading .env file: %v", err)
	}

	redisClient := redisdb.Client
	ctx := context.Background()

	// Use SSCAN for batch processing
	var cursor uint64
	uniqueMap := make(map[string]string)
	batchSize := int64(1000) // adjust if needed

	for {
		// Use SSCAN to iterate over Redis set in batches
		values, newCursor, err := redisClient.SScan(ctx, "products:crawler", cursor, "*", batchSize).Result()
		if err != nil {
			log.Fatalf("❌ SSCAN failed: %v", err)
		}

		for _, item := range values {
			var product Crawler.Product
			err := json.Unmarshal([]byte(item), &product)
			if err != nil {
				log.Printf("⚠️ Failed to parse product: %v", err)
				continue
			}

			// Create a unique key for each product based on title, marketplace_id, and seller_name
			key := fmt.Sprintf("%s_%d_%s", product.Title, product.MarketplaceID, product.SellerName)
			if _, exists := uniqueMap[key]; !exists {
				uniqueMap[key] = item
			}
		}

		// Update cursor to continue scanning
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	// Convert the unique products into a slice of Product objects
	var uniqueProducts []Crawler.Product
	for _, item := range uniqueMap {
		var product Crawler.Product
		err := json.Unmarshal([]byte(item), &product)
		if err != nil {
			log.Printf("⚠️ Failed to parse unique product: %v", err)
			continue
		}

		var weight *float32
		if product.Weight != nil {
			w := *product.Weight
			weight = &w
		} else {
			defaultWeight := float32(0)
			weight = &defaultWeight
		}

		// Convert to models.Product (with your fields)
		// After deduplicating, map the Redis product to the PostgreSQL model:
		uniqueProducts = append(uniqueProducts, Crawler.Product{
			Title: product.Title,
			Brand: product.Brand, // This is an optional field, make sure to add it
			Image: Crawler.Image{
				Small: product.Image.Small,
				Large: product.Image.Large,
			},
			Price:         product.Price,
			OriginalPrice: product.OriginalPrice,
			Discount:      product.Discount,
			Rating:        product.Rating,
			RatingCount:   product.RatingCount,
			Sold:          product.Sold,
			SellerName:    product.SellerName,
			Location: Crawler.Location{
				City: product.Location.City,
			},
			MarketplaceID: product.MarketplaceID,
			KeywordID:     product.KeywordID,
			ComodityID:    product.ComodityID,
			Description:   product.Description,
			Link:          product.Link,
			Weight:        weight,
			Created_at:    product.Created_at,
		})

	}

	// Save unique products to PostgreSQL
	err := Crawler.SaveProductsToPostgres(uniqueProducts)
	if err != nil {
		log.Fatalf("❌ Failed to save products to PostgreSQL: %v", err)
	}

	fmt.Println("✅ Deduplication and saving to PostgreSQL complete!")
}
