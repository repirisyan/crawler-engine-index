package models

import (
	"context"
	pgdb "crawler-index/db/postgres"
	"log"
	"time"
)

var Ctx = context.Background() // Ensure you have a context here

type RegionBrandLeaderboard struct {
	Brand_id             uint64 `json:"brand_id"`
	Brand_leaderboard_id uint64 `json:"brand_leaderboard_id"`
}

func SaveBrandToPostgres(brands []RegionBrandLeaderboard) error {
	// Get PostgreSQL connection pool
	conn := pgdb.GetPostgresPool()

	// Begin a transaction
	tx, err := conn.Begin(Ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return err
	}

	query := `INSERT INTO region_brand_leader_boards (brand_id, brand_leader_board_id, created_at, updated_at) VALUES ($1, $2, $3, $4)`

	for _, brand := range brands {
		_, err = tx.Exec(Ctx, query, brand.Brand_id, brand.Brand_leaderboard_id, time.Now(), time.Now())

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

	log.Println("✅ Products saved to PostgreSQL")
	return nil
}
