// models/mysql/storeTempItem.go
package mysqlTempItem

import (
	"crawler-index/db/mysql"
	"fmt"
)

type Product struct {
	Title          string 
	Link           string 
	Image          *string 
	Price          uint64 
	Rating         float64
	Sold           uint64 
	Seller         string 
	Location       string 
	Comodity_id    uint64 
	Keyword_id     uint64 
	Marketplace_id uint64
	User_id 	   uint64  
	Created_at     string
}

type Supervision struct {
	Supervision_id uint64
	Title          string
	Link           string
	Image          *string
	Price          uint64
	Sold           uint64
	Seller         string
	Location       string
	Comodity_id    uint64
	Marketplace_id uint64
	Created_at     string
}

func GetAllProduct(offset, limit int, search string) ([]Supervision, error) {
	query := "SELECT id, title, link, image, price, sold, seller, location, comodity_id, marketplace_id from temp_items WHERE title LIKE ? LIMIT ?, ?"
	likeSearch := "%" + search + "%"
	rows, err := mysql.DB.Query(query,likeSearch, offset, limit)
	if err != nil {
		return nil, err
	}

	var products []Supervision

	for rows.Next() {
		var product Supervision
		err := rows.Scan(&product.Supervision_id,&product.Title, &product.Link, &product.Image, &product.Price, &product.Sold, &product.Seller, &product.Location, &product.Comodity_id, &product.Marketplace_id)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func GetAllData(search string) ([]Supervision, error) {
	query := "SELECT title, link, image, price, sold, seller, location, comodity_id, marketplace_id, created_at from temp_items WHERE title LIKE ?"
	likeSearch := "%" + search + "%"
	rows, err := mysql.DB.Query(query,likeSearch)
	if err != nil {
		return nil, err
	}

	var products []Supervision

	for rows.Next() {
		var product Supervision
		err := rows.Scan(&product.Title, &product.Link, &product.Image, &product.Price, &product.Sold, &product.Seller, &product.Location, &product.Comodity_id, &product.Marketplace_id,&product.Created_at)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// StoreProduct inserts a product into the database
func StoreProduct(product Product) error {
	query := "INSERT INTO temp_items (title, link, image, price, rating, sold, seller, location, comodity_id, keyword_id, marketplace_id, user_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := mysql.DB.Exec(query, product.Title, product.Link, product.Image, product.Price, product.Rating, product.Sold, product.Seller, product.Location, product.Comodity_id, product.Keyword_id, product.Marketplace_id, product.User_id, product.Created_at)
	if err != nil {
		return fmt.Errorf("error inserting product: %v", err)
	}
	return nil
}

func UpdateFlag(title string,seller string, marketplace_id uint64) error {
	query := "UPDATE temp_items SET FLAG = 1 WHERE title = ? AND seller = ? AND marketplace_id = ?"
	_, err := mysql.DB.Exec(query, title, seller, marketplace_id)
	if err != nil {
		return fmt.Errorf("error update product: %v", err)
	}
	return nil
}
