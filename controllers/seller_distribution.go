package controllers

import (
	"fmt"
	"log"
	"time"

	Crawler "crawler-index/models/postgres/crawler"
	SellerDistribution "crawler-index/models/postgres/seller_distribution"
)

// Store Data from temp_item to products in mongodb
func StoreSellerDistribution() {
	fmt.Println("Store Seller Distribution Data")

	results, err := Crawler.GetSellerDistribution()

	if err != nil {
		fmt.Printf("Error fetching seller distribution: %v\n", err)
	}
	// Convert struct to interface
	var sellerDistributionResult []SellerDistribution.SellerDistribution
	for _, p := range results {
		sellerDistributionResult = append(sellerDistributionResult, SellerDistribution.SellerDistribution{
			Comodity_id:    p.Comodity_id,
			Marketplace_id: p.Marketplace_id,
			Seller:         p.Total,
			Year:           time.Now().Year(),
			Month:          uint8(time.Now().Month()),
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
