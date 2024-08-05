// models/mysql/storeTempItem.go
package mysqlSupervision

import (
	"crawler-index/db/mysql"
)

type Supervision struct {
	Name           string
	Link           string
	Image          *string
	Price          uint64
	Sold           uint64
	Seller         string
	Location       string
	Keyword_id     uint64
	Marketplace_id uint64
	Created_at	   string
}

// StoreProduct inserts a product into the database
func StoreSupervisions(supervisions []Supervision) error {
	query := "INSERT INTO supervisions (name, link, image, keyword_id, price, marketplace_id, seller, location, sold, created_at) VALUES"
	values := []interface{}{}

	for _, supervision := range supervisions {
        query += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
        values = append(values, supervision.Name, supervision.Link, supervision.Image, supervision.Keyword_id, supervision.Price, supervision.Marketplace_id, supervision.Seller, supervision.Location, supervision.Sold, supervision.Created_at)
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
