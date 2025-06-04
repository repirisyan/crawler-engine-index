package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	pgdb "crawler-index/db/postgres"
)

var Ctx = context.Background() // Ensure you have a context here

type Image struct {
	Small []string `json:"small"`
	Large []string `json:"large"`
}

type Location struct {
	Province *string `json:"province,omitempty"`
	City     *string `json:"city,omitempty"`
}

type Product struct {
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
	Bpom                bool           `json:"bpom"`
	Bpom_number         sql.NullString `json:"bpom_number,omitempty"`
	Sni                 bool           `json:"sni"`
	Halal               bool           `json:"halal"`
	Distribution_permit bool           `json:"distribution_permit"`
}

func StoreProductIndex(products []Product) error {
	// Get PostgreSQL connection pool
	conn := pgdb.GetPostgresPool()

	// Begin a transaction
	tx, err := conn.Begin(Ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
		return err
	}

	// Prepare the insert query (including created_at and updated_at)
	query := `INSERT INTO products (title, brand, price, original_price, discount, rating, rating_count, sold, seller_name, seller_url, location, weight, description, marketplace_id, comodity_id, keyword_id, link, image, bpom, bpom_number, sni, halal, distribution_permit, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24) ON CONFLICT (title, seller_name, marketplace_id) DO NOTHING;`

	// Insert each product into the database
	for _, product := range products {
		// Marshal Location and Image to JSON
		locationJSON, err := json.Marshal(product.Location)
		if err != nil {
			log.Printf("⚠️ Failed to marshal location: %v", err)
			continue
		}

		imageJSON, err := json.Marshal(product.Image)
		if err != nil {
			log.Printf("⚠️ Failed to marshal image: %v", err)
			continue
		}

		// Execute the query with all columns, including created_at and updated_at
		_, err = tx.Exec(Ctx, query,
			product.Title,         // title
			product.Brand,         // brand (nullable)
			product.Price,         // price
			product.OriginalPrice, // original_price
			product.Discount,      // discount
			product.Rating,        // rating
			product.RatingCount,   // rating_count
			product.Sold,          // sold
			product.SellerName,    // seller_name
			product.SellerURL,     // seller_url (nullable)
			locationJSON,          // location (JSON)
			product.Weight,        // weight
			product.Description,   // description
			product.MarketplaceID, // marketplace_id
			product.ComodityID,    // commodity_id
			product.KeywordID,     // keyword_id
			product.Link,          // link
			imageJSON,             // image (JSON)
			product.Bpom,
			product.Bpom_number,
			product.Sni,
			product.Halal,
			product.Distribution_permit,
			time.Now(),
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

	log.Println("✅ Products Index saved to PostgreSQL")
	return nil
}
