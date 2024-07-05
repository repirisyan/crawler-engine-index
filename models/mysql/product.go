// models/mysql/product.go
package product

import (
    "crawler-index/db/mysql"
)

type Product struct {
    Title string
    Image *string
    Price uint64
    Rating float32
    Sold uint32
    Seller string
    Location string
    Comodity string
    Marketplace string
    Created_at string
}

// GetAllUsers retrieves all products from the database
func GetAllProduct(offset, limit int) ([]Product, error) {
    query := "SELECT p.title, p.image, p.price, p.rating, p.sold, p.seller, p.location, p.created_at, c.name as comodity, m.name as marketplace FROM temp_items as p JOIN comodities as c ON c.id = p.comodity_id JOIN marketplaces as m ON m.id = p.marketplace_id LIMIT ?, ?"
    rows, err := mysql.DB.Query(query, offset, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var products []Product

    for rows.Next() {
        var product Product
        err := rows.Scan(&product.Title, &product.Image, &product.Price, &product.Rating, &product.Sold, &product.Seller, &product.Location, &product.Created_at, &product.Comodity, &product.Marketplace)
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
