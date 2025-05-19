package models

import (
	"context"
	"log"

	pgdb "crawler-index/db/postgres"
)

var Ctx = context.Background() // Ensure you have a context here

type SellerDistribution struct {
	Comodity_id    uint64 `bson:"comodity_id"`
	Marketplace_id uint64 `bson:"marketplace_id"`
	Seller         uint32 `bson:"total_seller"`
	Year           int    `bson:"year"`
	Month          uint8  `bson:"month"`
}

func StoreSellerDistribution(sellerDistributions []SellerDistribution) error {
	// Get PostgreSQL connection pool
	conn := pgdb.GetPostgresPool()

	// Begin a transaction
	tx, err := conn.Begin(Ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
		return err
	}

	query := `INSERT INTO seller_distribution_summaries (comodity_id, marketplace_id, total_seller, year, month)
				VALUES ($1, $2, $3, $4, $5)`

	// Insert each product into the database
	for _, seller_distribution := range sellerDistributions {
		// Execute the query with all columns, including created_at and updated_at
		_, err = tx.Exec(Ctx, query,
			seller_distribution.Comodity_id,
			seller_distribution.Marketplace_id,
			seller_distribution.Seller,
			seller_distribution.Year,
			seller_distribution.Month,
		)
		if err != nil {
			tx.Rollback(Ctx) // Rollback transaction if any error occurs
			log.Printf("Failed to insert product: %v", err)
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(Ctx); err != nil {
		tx.Rollback(Ctx)
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	log.Println("✅ Seller Distribution saved to PostgreSQL")
	return nil

}
