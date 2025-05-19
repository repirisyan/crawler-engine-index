package models

import (
	"context"
	"log"

	pgdb "crawler-index/db/postgres"
)

var Ctx = context.Background() // Ensure you have a context here

type BrandLeaderboard struct {
	Brand          string `bson:"brand"`
	Marketplace_id uint64 `json:"marketplace_id"`
	Total          uint32 `bson:"total"`
	Year           int    `bson:"year"`
	Month          uint8  `bson:"month"`
}

func StoreBrandLeaderboard(brands []BrandLeaderboard) error {
	// Get PostgreSQL connection pool
	conn := pgdb.GetPostgresPool()

	// Begin a transaction
	tx, err := conn.Begin(Ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
		return err
	}

	// Prepare the insert query (including created_at and updated_at)
	query := `INSERT INTO brand_leader_boards (brand, marketplace_id, total_product, year, month)
				VALUES ($1, $2, $3, $4, $5)`

	// Insert each product into the database
	for _, brand := range brands {
		_, err = tx.Exec(Ctx, query,
			brand.Brand,
			brand.Marketplace_id,
			brand.Total,
			brand.Year,
			brand.Month,
		)
		if err != nil {
			tx.Rollback(Ctx) // Rollback transaction if any error occurs
			log.Printf("Failed to insert brand: %v", err)
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(Ctx); err != nil {
		tx.Rollback(Ctx)
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	log.Println("✅ Products saved to PostgreSQL")
	return nil
}
