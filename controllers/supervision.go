package controllers

import (
	"fmt"
	"log"

	Crawler "crawler-index/models/postgres/crawler"
	Supervision "crawler-index/models/postgres/supervision"
)

// Store Data from temp_item to indexing data
func StoreSupervisionData() {
	fmt.Println("Store Supervision Data")

	limit := 1000

	supervisions, err := Supervision.GetAllSupervisionList()
	if err != nil {
		fmt.Printf("Error fetching supervision list: %v\n", err)
		return
	}

	for _, supervision := range supervisions {
		offset := 0
		for {
			products, err := Crawler.SearchProduct(offset, limit, supervision.Name)
			if err != nil {
				fmt.Printf("Error fetching products: %v\n", err)
				break
			}
			var productResult []Supervision.Supervision
			for _, product := range products {
				productResult = append(productResult, Supervision.Supervision{
					ProductID: product.ID,
					Title:     product.Title,
					Brand:     product.Brand, // This is an optional field, make sure to add it
					Image: Supervision.Image{
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
					Location: Supervision.Location{
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
					Crawled_at:          product.Created_at,
					Label:               supervision.Category,
				})
			}

			if len(productResult) > 0 {
				productID, err := Supervision.SaveSupervisionsToPostgres(productResult)
				if err != nil {
					log.Printf("Failed to insert batch supervision product: %v\n", err)
					continue
				}
				err = Crawler.SetSupervised(productID)
			}

			if len(products) == 0 {
				// No more records to fetch
				break
			}

			// Update offset for next iteration
			offset += limit
		}
	}
	log.Printf("Failed to insert batch supervision product: %v\n", err)

}
