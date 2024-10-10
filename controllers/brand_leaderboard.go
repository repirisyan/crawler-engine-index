package controllers

import (
	"fmt"
	"log"
	"time"

	BrandLeaderboard "crawler-index/models/mongodb/brand_leaderboard"
	MongoTempItem "crawler-index/models/mongodb/temp_item"
)

// Store Data from temp_item to products in mongodb
func StoreBrandLeaderboard() {
	fmt.Println("Store Brand Leaderboard Data")

	results, err := MongoTempItem.GetBrandLeaderboard()

	if err != nil {
		fmt.Printf("Error fetching brand leaderboard: %v\n", err)
	}
	// Convert struct to interface
	var brandLeaderBoardResult []interface{}
	for _, p := range results {
		brandLeaderBoardResult = append(brandLeaderBoardResult, BrandLeaderboard.BrandLeaderboard{
			Brand:    p.Brand,
			Marketplace: p.Marketplace,
			Total:      p.Total,
			Year:        time.Now().Year(),
			Month:       int(time.Now().Month()),
			Date:        time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	if len(brandLeaderBoardResult) > 0 {
		err = BrandLeaderboard.StoreBrandLeaderboard(brandLeaderBoardResult)
		if err != nil {
			log.Printf("Failed to insert batch products: %v\n", err)
		}
	}

	fmt.Println("Store Brand Leaderboard Data Done")

}
