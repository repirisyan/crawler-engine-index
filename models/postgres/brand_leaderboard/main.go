package models

import (
	"context"
	pgdb "crawler-index/db/postgres"
	"log"
)

var Ctx = context.Background() // Ensure you have a context here

type BrandLeaderboard struct {
	Brand          string `bson:"brand"`
	Marketplace_id uint64 `json:"marketplace_id"`
	Comodity_id    uint64 `json:"comodity_id"`
	Total          uint32 `bson:"total"`
	Year           int    `bson:"year"`
	Month          uint8  `bson:"month"`
}

type BrandID struct {
	ID uint64 `json:"id"`
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
	query := `INSERT INTO brand_leader_boards (brand, marketplace_id,comodity_id, total_product, year, month)
				VALUES ($1, $2, $3, $4, $5, $6)`

	// Insert each product into the database
	for _, brand := range brands {
		_, err = tx.Exec(Ctx, query,
			brand.Brand,
			brand.Marketplace_id,
			brand.Comodity_id,
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

func SearchBrand(brand string) ([]BrandID, error) {
	conn := pgdb.GetPostgresPool()

	query := `SELECT id FROM brand_leader_boards WHERE brand = $1`

	rows, err := conn.Query(Ctx, query, brand)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var brands []BrandID

	for rows.Next() {
		var brand BrandID
		err := rows.Scan(
			&brand.ID,
		)
		if err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return brands, nil
}
