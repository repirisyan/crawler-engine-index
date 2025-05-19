package controllers

import (
	"fmt"
	"log"
	"time"

	Crawler "crawler-index/models/postgres/crawler"
	DiscountProduct "crawler-index/models/postgres/discount_product"
)

// Store Data from temp_item to products in mongodb
func StoreDiscountProduct() {
	fmt.Println("Store Discount Product Data Start")

	results, err := Crawler.GetDiscountProduct()

	if err != nil {
		fmt.Printf("Error fetching Discount Product: %v\n", err)
	}
	// Convert struct to interface
	var discountProductResult []DiscountProduct.DiscountSummaries
	for _, p := range results {
		discountProductResult = append(discountProductResult, DiscountProduct.DiscountSummaries{
			Discount:       p.Discount,
			Marketplace_id: p.Marketplace_id,
			Total:          p.Total,
			Year:           time.Now().Year(),
			Month:          uint8(time.Now().Month()),
		})
	}

	if len(discountProductResult) > 0 {
		err = DiscountProduct.StoreDiscountSummaries(discountProductResult)
		if err != nil {
			log.Printf("Failed to insert batch discount products: %v\n", err)
		}
	}

	fmt.Println("Store Discount Product Data Done")

}
