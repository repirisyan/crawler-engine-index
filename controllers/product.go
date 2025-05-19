package controllers

import (
	"fmt"
	"log"

	Crawler "crawler-index/models/postgres/crawler"
	ProductIndex "crawler-index/models/postgres/product_index"
)

// Store Data from temp_item to products in mongodb
func StoreIndexData() {
	fmt.Println("Store Index Data")

	limit := 1000
	offset := 0
	for {
		products, err := Crawler.GetAllProducts(offset, limit)
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}
		// Convert struct to interface
		var productResult []ProductIndex.Product
		for _, product := range products {
			productResult = append(productResult, ProductIndex.Product{
				Title: product.Title,
				Brand: product.Brand, // This is an optional field, make sure to add it
				Image: ProductIndex.Image{
					Small: product.Image.Small,
					Large: product.Image.Large,
				},
				Price:         product.Price,
				OriginalPrice: product.OriginalPrice,
				Discount:      product.Discount,
				Rating:        product.Rating,
				RatingCount:   product.RatingCount,
				Sold:          product.Sold,
				SellerName:    product.SellerName,
				SellerURL:     product.SellerURL,
				Location: ProductIndex.Location{
					City: product.Location.City,
				},
				MarketplaceID:       product.MarketplaceID,
				KeywordID:           product.KeywordID,
				ComodityID:          product.ComodityID,
				Description:         product.Description,
				Link:                product.Link,
				Weight:              product.Weight,
				Bpom:                product.Bpom,
				Bpom_number:         product.Bpom_number,
				Sni:                 product.Sni,
				Halal:               product.Halal,
				Distribution_permit: product.Distribution_permit,
			})
		}

		if len(productResult) > 0 {
			err = ProductIndex.StoreProductIndex(productResult)
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
