package controllers

import (
	"fmt"
	"log"
	"time"

	SellerDistribution "crawler-index/models/mongodb/seller_distribution"
	MongoTempItem "crawler-index/models/mongodb/temp_item"
)

// Store Data from temp_item to products in mongodb
func StoreSellerDistribution() {
	fmt.Println("Store Seller Distribution Data")

	results, err := MongoTempItem.GetSellerDistribution()

	if err != nil {
		fmt.Printf("Error fetching seller distribution: %v\n", err)
	}
	// Convert struct to interface
	var sellerDistributionResult []interface{}
	for _, p := range results {
		sellerDistributionResult = append(sellerDistributionResult, SellerDistribution.SellerDistribution{
			Comodity:    p.Comodity,
			Marketplace: p.Marketplace,
			Seller:      p.Total,
			Year:        time.Now().Year(),
			Month:       int(time.Now().Month()),
			Date:        time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	if len(sellerDistributionResult) > 0 {
		err = SellerDistribution.StoreSellerDistribution(sellerDistributionResult)
		if err != nil {
			log.Printf("Failed to insert batch products: %v\n", err)
		}
	}

	fmt.Println("Store Seller Distribution Data Done")

}
