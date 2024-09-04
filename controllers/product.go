package controllers

import (
	"fmt"
	"log"

	"crawler-index/models/mongodb/product"
	"crawler-index/models/mongodb/temp_item"
)

// Store Data from temp_item to products in mongodb
func StoreIndexData() {
	fmt.Println("Store Index Data")

	limit := 1000
	offset := 0
	for {
		products, err := MongoTempItem.GetAllProducts(offset, limit)
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}
		// Convert struct to interface
		var productResult []interface{}
		for _, p := range products {
			productResult = append(productResult, MongoProduct.IndexProduct{
				Title:       p.Title,
				Description: p.Description,
				Category:    p.Category,
				Link:        p.Link,
				Image:       p.Image,
				Price:       p.Price,
				Rating:      p.Rating,
				Sold:        p.Sold,
				Seller:      p.Seller,
				Location:    p.Location,
				Certified:   p.Certified,
				Comodity:    p.Comodity,
				Keyword:     p.Keyword,
				Marketplace: p.Marketplace,
				Created_at:  p.Created_at,
			})
		}

		if len(productResult) > 0 {
			err = MongoProduct.StoreProducts(productResult)
			if err != nil {
				log.Printf("Failed to insert batch products: %v\n", err)
				continue
			}
		}

		if len(products) == 0 {
			// No more records to fetch
			break
		}

		// Update offset for next iteration
		offset += limit
	}
}
