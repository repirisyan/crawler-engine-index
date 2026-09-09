package controllers

import (
	"fmt"
	"log"
	"time"

	Brand "crawler-index/models/postgres/brand"
	BrandLeaderboard "crawler-index/models/postgres/brand_leaderboard"
	RegionBrand "crawler-index/models/postgres/region_brand_leaderboard"
)

func StoreRegionBrand() {
	fmt.Println("Store Region Brand Data")

	now := time.Now()
	year, month := now.Year(), int(now.Month())

	limit := 1000
	offset := 0

	for {
		brands, err := Brand.GetAllBrand(offset, limit)

		if err != nil {
			fmt.Printf("Error fetching brands list: %v\n", err)
			return
		}
		var brandResult []RegionBrand.RegionBrandLeaderboard

		for _, brand := range brands {
			brand_results, err := BrandLeaderboard.SearchBrand(brand.Name, year, month)

			if err != nil {
				fmt.Printf("Error fetching brand leaderboard: %v\n", err)
				break
			}

			for _, brand_result := range brand_results {
				brandResult = append(brandResult, RegionBrand.RegionBrandLeaderboard{
					Brand_id:             brand.ID,
					Brand_leaderboard_id: brand_result.ID,
				})
			}
		}

		if len(brandResult) > 0 {
			err := RegionBrand.SaveBrandToPostgres(brandResult)
			if err != nil {
				log.Printf("Failed to insert batch brand region: %v\n", err)
				continue
			}
		}

		if len(brandResult) == 0 {
			// No more records to fetch
			break
		}

		// Update offset for next iteration
		offset += limit
	}
}
