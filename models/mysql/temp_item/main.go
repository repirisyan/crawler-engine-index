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
	Keyword_id     uint64
	Marketplace_id uint64
	Created_at     string
}

type FlagUpdate struct {
	Title string
	Seller string
	Marketplace_id uint64
}

func GetAllProduct(offset, limit int, search string) ([]Supervision, error) {
	query := "SELECT id, title, link, image, price, sold, seller, location, keyword_id, marketplace_id from temp_items WHERE title LIKE ? LIMIT ?, ?"
	likeSearch := "%" + search + "%"
	rows, err := mysql.DB.Query(query,likeSearch, offset, limit)
	if err != nil {
		return nil, err
	}

	var products []Supervision

	for rows.Next() {
		var product Supervision
		err := rows.Scan(&product.Supervision_id,&product.Title, &product.Link, &product.Image, &product.Price, &product.Sold, &product.Seller, &product.Location, &product.Keyword_id, &product.Marketplace_id)
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
	query := "SELECT title, link, image, price, sold, seller, location, keyword_id, marketplace_id, created_at from temp_items WHERE title LIKE ?"
	likeSearch := "%" + search + "%"
	rows, err := mysql.DB.Query(query,likeSearch)
	if err != nil {
		return nil, err
	}

	var products []Supervision

	for rows.Next() {
		var product Supervision
		err := rows.Scan(&product.Title, &product.Link, &product.Image, &product.Price, &product.Sold, &product.Seller, &product.Location, &product.Keyword_id, &product.Marketplace_id,&product.Created_at)
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
func StoreProducts(products []Product) error {
	query := "INSERT INTO temp_items (Title, Link, Image, Price, Rating, Sold, Seller, Location, Keyword_id, Marketplace_id, User_id, Created_at) VALUES "
	values := []interface{}{}

    for _, product := range products {
        query += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
        values = append(values, product.Title, product.Link, product.Image, product.Price, product.Rating, product.Sold, product.Seller, product.Location, product.Keyword_id, product.Marketplace_id, product.User_id, product.Created_at)
    }

    query = query[:len(query)-1] // Remove the trailing comma

    stmt, err := mysql.DB.Prepare(query)
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(values...)
    if err != nil {
        return err
    }

    return nil
}

func UpdateFlags(flagUpdates []FlagUpdate) error {
	query := "UPDATE temp_items SET flag = 1 WHERE (Title, Seller, Marketplace_id) IN ("
    values := []interface{}{}

    for _, update := range flagUpdates {
        query += "(?, ?, ?),"
        values = append(values, update.Title, update.Seller, update.Marketplace_id)
    }

    query = query[:len(query)-1] // Remove the trailing comma
    query += ")"

    stmt, err := mysql.DB.Prepare(query)
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(values...)
    if err != nil {
        return err
    }

    return nil
}
