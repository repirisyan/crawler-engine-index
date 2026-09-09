package models

import (
	"context"
	pgdb "crawler-index/db/postgres"
	"crawler-index/models/postgres/dbutil"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

var Ctx = context.Background() // Ensure you have a context here

type SupervisionList struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Image struct {
	Small []string `json:"small"`
	Large []string `json:"large"`
}

type Location struct {
	Province *string `json:"province,omitempty"`
	City     *string `json:"city,omitempty"`
}

type Supervision struct {
	ProductID           uint64         `json:"id,omitempty"`
	Title               string         `json:"title"`
	Brand               *string        `json:"brand,omitempty"`
	Image               Image          `json:"image"`
	Price               uint64         `json:"price"`
	OriginalPrice       uint64         `json:"original_price"`
	Discount            float64        `json:"discount"`
	Rating              float64        `json:"rating"`
	RatingCount         uint64         `json:"rating_count"`
	Sold                uint64         `json:"sold"`
	SellerName          string         `json:"seller_name"`
	SellerURL           *string        `json:"seller_url,omitempty"`
	Location            Location       `json:"location"`
	MarketplaceID       uint64         `json:"marketplace_id"`
	KeywordID           uint64         `json:"keyword_id"`
	ComodityID          uint64         `json:"comodity_id"`
	Description         string         `json:"description,omitempty"`
	Link                string         `json:"link"`
	Weight              *float32       `json:"weight,omitempty"`
	Bpom                bool           `json:"bpom,omitempty"`
	Bpom_number         sql.NullString `json:"bpom_number,omitempty"`
	Sni                 bool           `json:"sni,omitempty"`
	Halal               bool           `json:"halal,omitempty"`
	Distribution_permit bool           `json:"distribution_permit,omitempty"`
	Crawled_at          time.Time      `json:"crawled_at"`
	Label               string         `json:"label"`
}

func GetAllSupervisionList() ([]SupervisionList, error) {
	conn := pgdb.GetPostgresPool()

	query := `SELECT name, category FROM supervision_lists WHERE status = true`

	rows, err := conn.Query(Ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var supervision_lists []SupervisionList
	for rows.Next() {
		var supervision_list SupervisionList
		err := rows.Scan(
			&supervision_list.Name,
			&supervision_list.Category,
		)
		if err != nil {
			return nil, err
		}
		supervision_lists = append(supervision_lists, supervision_list)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return supervision_lists, nil
}

func SaveSupervisionsToPostgres(supervisions []Supervision) ([]uint64, error) {
	// Get PostgreSQL connection pool
	conn := pgdb.GetPostgresPool()

	// Begin a transaction
	tx, err := conn.Begin(Ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback(Ctx) // no-op once committed

	// Prepare the insert query (including created_at and updated_at)
	query := `INSERT INTO supervisions (title, brand, price, original_price, discount, rating, rating_count, sold, seller_name, seller_url, location, weight, description, marketplace_id, comodity_id, keyword_id, link, image, created_at, updated_at, crawled_at, label)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) ON CONFLICT (title, seller_name, marketplace_id,date, description) DO NOTHING RETURNING id;`

	var inserted []uint64
	skipped := 0
	// Insert each product into the database
	for _, supervision := range supervisions {
		// Marshal Location and Image to JSON
		locationJSON, err := json.Marshal(supervision.Location)
		if err != nil {
			log.Printf("⚠️ Failed to marshal location: %v", err)
			skipped++
			continue
		}

		imageJSON, err := json.Marshal(supervision.Image)
		if err != nil {
			log.Printf("⚠️ Failed to marshal image: %v", err)
			skipped++
			continue
		}

		// One savepoint per row so a single bad row is skipped instead of
		// aborting the whole batch.
		sp, err := tx.Begin(Ctx)
		if err != nil {
			log.Printf("Failed to create savepoint: %v", err)
			return nil, err
		}

		var rowID uint64
		err = sp.QueryRow(Ctx, query,
			dbutil.Truncate(supervision.Title, 1000),   // title
			dbutil.TruncatePtr(supervision.Brand, 255), // brand (nullable)
			supervision.Price,                          // price
			supervision.OriginalPrice,                  // original_price
			supervision.Discount,                       // discount
			supervision.Rating,                         // rating
			supervision.RatingCount,                    // rating_count
			supervision.Sold,                           // sold
			dbutil.Truncate(supervision.SellerName, 1000),   // seller_name
			dbutil.TruncatePtr(supervision.SellerURL, 1000), // seller_url (nullable)
			locationJSON,                                    // location (JSON)
			supervision.Weight,                              // weight
			supervision.Description,                         // description
			supervision.MarketplaceID,                       // marketplace_id
			supervision.ComodityID,                          // commodity_id
			supervision.KeywordID,                           // keyword_id
			dbutil.Truncate(supervision.Link, 1500),         // link
			imageJSON,                                       // image (JSON)
			time.Now(),
			time.Now(),
			supervision.Crawled_at,
			supervision.Label,
		).Scan(&rowID)

		if err != nil {
			// ErrNoRows just means the row already existed (ON CONFLICT).
			if !errors.Is(err, pgx.ErrNoRows) {
				log.Printf("Skipping supervision %q: %v", supervision.Title, err)
				skipped++
			}
			sp.Rollback(Ctx)
			continue
		}
		if err := sp.Commit(Ctx); err != nil { // release savepoint
			log.Printf("Failed to release savepoint: %v", err)
			return nil, err
		}

		inserted = append(inserted, supervision.ProductID)
	}

	// Commit the transaction
	if err := tx.Commit(Ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return nil, err
	}

	if skipped > 0 {
		log.Printf("✅ Supervisions saved to PostgreSQL (%d row(s) skipped)", skipped)
	} else {
		log.Println("✅ Supervisions saved to PostgreSQL")
	}
	return inserted, nil
}
