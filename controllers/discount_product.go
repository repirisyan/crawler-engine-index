package controllers

import (
	"fmt"
	"log"
	"time"

	DiscountProduct "crawler-index/models/mongodb/discount_product"
	MongoTempItem "crawler-index/models/mongodb/temp_item"
)

// Store Data from temp_item to products in mongodb
func StoreDiscountProduct() {
	fmt.Println("Store Discount Product Data Start")

	results, err := MongoTempItem.GetDiscountProduct()

	if err != nil {
		fmt.Printf("Error fetching Discount Product: %v\n", err)
	}
	// Convert struct to interface
	var discountProductResult []interface{}
	for _, p := range results {
		discountProductResult = append(discountProductResult, DiscountProduct.DiscountProduct{
			Discount:    p.Discount,
			Marketplace: p.Marketplace,
			Total:      p.Total,
			Year:        time.Now().Year(),
			Month:       int(time.Now().Month()),
			Date:        time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	if len(discountProductResult) > 0 {
		err = DiscountProduct.StoreDiscountProduct(discountProductResult)
		if err != nil {
			log.Printf("Failed to insert batch discount products: %v\n", err)
		}
	}

	fmt.Println("Store Discount Product Data Done")

}
