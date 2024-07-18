// models/mysql/storeTempItem.go
package mysqlSupervision

import (
	"crawler-index/db/mysql"
	"fmt"
)

type Supervision struct {
	Name           string
	Link           string
	Image          *string
	Price          uint64
	Sold           uint64
	Seller         string
	Location       string
	Comodity_id    uint64
	Marketplace_id uint64
}

// StoreProduct inserts a product into the database
func StoreSupervision(supervision Supervision) error {
	query := "INSERT INTO supervisions (name, link, image, comodity_id, price, marketplace_id, seller, location, sold) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := mysql.DB.Exec(query, supervision.Name, supervision.Link, supervision.Image, supervision.Comodity_id, supervision.Price, supervision.Marketplace_id, supervision.Seller, supervision.Location, supervision.Sold)
	if err != nil {
		return fmt.Errorf("error inserting product: %v", err)
	}
	return nil
}
