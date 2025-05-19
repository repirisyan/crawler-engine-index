package controllers

import (
	"fmt"
	"log"
	"time"

	BrandLeaderboard "crawler-index/models/postgres/brand_leaderboard"
	Crawler "crawler-index/models/postgres/crawler"
)

// Store Data from temp_item to products in mongodb
func StoreBrandLeaderboard() {
	fmt.Println("Store Brand Leaderboard Data")

	results, err := Crawler.GetBrandLeaderboard()

	if err != nil {
		fmt.Printf("Error fetching brand leaderboard: %v\n", err)
	}
	// Convert struct to interface
	var brandLeaderBoardResult []BrandLeaderboard.BrandLeaderboard
	for _, p := range results {
		brandLeaderBoardResult = append(brandLeaderBoardResult, BrandLeaderboard.BrandLeaderboard{
			Brand:          p.Brand,
			Marketplace_id: p.Marketplace_id,
			Total:          p.Total,
			Year:           time.Now().Year(),
			Month:          uint8(time.Now().Month()),
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
