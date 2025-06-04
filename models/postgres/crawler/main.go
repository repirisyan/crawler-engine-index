package models

import (
	"context"
	pgdb "crawler-index/db/postgres"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
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
	ID                  uint64         `json:"id,omitempty"`
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
	Created_at          time.Time         `json:"created_at"`
}

type Certified struct {
	ID                  uint64
	Bpom                bool
	Bpom_number         string
	Sni                 bool
	Halal               bool
	Distribution_permit bool
}

type BrandLeaderboard struct {
	Brand          string `json:"brand"`
	Marketplace_id uint64 `json:"marketplace_id"`
	Total          uint32 `json:"total"`
}

type DiscountProduct struct {
	Discount       float32 `bson:"discount"`
	Marketplace_id uint64  `bson:"marketplace"`
	Total          uint32  `bson:"total"`
}

type SellerDistributionGroupedResult struct {
	Comodity_id    uint64 `bson:"comodity_id"`
	Marketplace_id uint64 `bson:"marketplace_id"`
	Total          uint32 `bson:"total"`
}

func SaveProductsToPostgres(products []Product) error {
	// Get PostgreSQL connection pool
	conn := pgdb.GetPostgresPool()

	// Begin a transaction
	tx, err := conn.Begin(Ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
		return err
	}

	// Prepare the insert query (including created_at and updated_at)
	query := `INSERT INTO crawlers (title, brand, price, original_price, discount, rating, rating_count, sold, seller_name, seller_url, location, weight, description, marketplace_id, comodity_id, keyword_id, link, image, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) ON CONFLICT (title, seller_name, marketplace_id) DO NOTHING;`

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
			imageJSON,             // image (JSON),
			product.Created_at,
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

	log.Println("✅ Products saved to PostgreSQL")
	return nil
}

func GetAllProducts(offset int, limit int) ([]Product, error) {
	conn := pgdb.GetPostgresPool()

	query := `SELECT id, title, brand, image, price, original_price, discount, rating, rating_count, sold,
       seller_name, seller_url, location, marketplace_id, keyword_id, comodity_id,
       description, link, weight, COALESCE(bpom, false), bpom_number, COALESCE(sni, false), COALESCE(halal, false), COALESCE(distribution_permit, false), created_at FROM crawlers LIMIT $2 OFFSET $1`

	rows, err := conn.Query(Ctx, query, offset, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ID,
			&product.Title,
			&product.Brand,
			&product.Image,
			&product.Price,
			&product.OriginalPrice,
			&product.Discount,
			&product.Rating,
			&product.RatingCount,
			&product.Sold,
			&product.SellerName,
			&product.SellerURL,
			&product.Location,
			&product.MarketplaceID,
			&product.KeywordID,
			&product.ComodityID,
			&product.Description,
			&product.Link,
			&product.Weight,
			&product.Bpom,
			&product.Bpom_number,
			&product.Sni,
			&product.Halal,
			&product.Distribution_permit,
			&product.Created_at,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func GetAllProductCertifications(offset int, limit int) ([]Product, error) {
	conn := pgdb.GetPostgresPool()

	query := `SELECT id, description FROM crawlers LIMIT $2 OFFSET $1`

	rows, err := conn.Query(Ctx, query, offset, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ID,
			&product.Description,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func SetCertified(productResult []interface{}) error {
	conn := pgdb.GetPostgresPool()

	query := `UPDATE crawlers SET
			bpom = $2,
			bpom_number = $3,
			sni = $4,
			halal = $5,
			distribution_permit = $6
		WHERE id = $1`

	for _, item := range productResult {
		product, ok := item.(Certified)
		if !ok {
			return errors.New("failed to cast item to Certified")
		}

		_, err := conn.Exec(Ctx, query,
			product.ID,
			product.Bpom,
			product.Bpom_number,
			product.Sni,
			product.Halal,
			product.Distribution_permit,
		)

		if err != nil {
			log.Printf("failed to update product: %v", err)
			return err
		}
	}

	return nil
}

func GetBrandLeaderboard() ([]BrandLeaderboard, error) {
	conn := pgdb.GetPostgresPool()

	// SQL query to aggregate the leaderboard by brand and marketplace
	query := `
		SELECT brand, marketplace_id, COUNT(*) as total
		FROM crawlers
		WHERE brand IS NOT NULL AND brand != ''
		GROUP BY brand, marketplace_id
		ORDER BY total DESC
	`

	// Perform the query
	rows, err := conn.Query(Ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error during query: %v", err)
	}
	defer rows.Close()

	var results []BrandLeaderboard
	batchCount := 0

	// Iterate over the rows
	for rows.Next() {
		var brand string
		var marketplace_id uint64
		var total uint32

		// Scan the values into variables
		if err := rows.Scan(&brand, &marketplace_id, &total); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Append the result to the slice
		results = append(results, BrandLeaderboard{
			Brand:          brand,
			Marketplace_id: marketplace_id,
			Total:          total,
		})

		// Increment batch counter and print progress
		if len(results)%1000 == 0 { // Adjust based on your batch size
			batchCount++
			fmt.Printf("Processed %d batches\n", batchCount)
		}
	}

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return results, nil
}

func GetDiscountProduct() ([]DiscountProduct, error) {
	conn := pgdb.GetPostgresPool()

	query := `
		SELECT 
			discount, 
			marketplace_id, 
			COUNT(*) AS total
		FROM 
			crawlers
		WHERE 
			discount IS NOT NULL 
			AND discount > 0
		GROUP BY 
			discount, marketplace_id;
	`

	rows, err := conn.Query(Ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []DiscountProduct
	for rows.Next() {
		var marketplace_id uint64
		var total uint32

		var discountVal float32
		if err := rows.Scan(&discountVal, &marketplace_id, &total); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		results = append(results, DiscountProduct{
			Discount:       discountVal,
			Marketplace_id: marketplace_id,
			Total:          total,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

func GetSellerDistribution() ([]SellerDistributionGroupedResult, error) {
	conn := pgdb.GetPostgresPool()

	query := `
		SELECT 
			comodity_id, 
			marketplace_id, 
			COUNT(*) AS total
		FROM 
			crawlers
		GROUP BY 
			comodity_id, marketplace_id;
	`

	rows, err := conn.Query(Ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var results []SellerDistributionGroupedResult

	for rows.Next() {
		var comodity_id uint64
		var marketplace_id uint64
		var total uint32

		if err := rows.Scan(&comodity_id, &marketplace_id, &total); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		results = append(results, SellerDistributionGroupedResult{
			Comodity_id:    comodity_id,
			Marketplace_id: marketplace_id,
			Total:          total,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}
