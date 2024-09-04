package controllers

import (
	"fmt"
	"log"
	"time"

	"crawler-index/models/mongodb/temp_item"
	"crawler-index/models/mongodb/train_data"
)

// Training Data for Machine Learning label Category
func StoreTrainingData() {
	fmt.Println("Store Training Data")

	limit := 1000
	offset := 0
	for {
		products, err := MongoTempItem.GetDataForTraining(offset, limit)
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}
		// Convert struct to interface
		var productResult []interface{}
		for _, p := range products {
			productResult = append(productResult, MongoTrainingData.TrainingData{
				Product_title: p.Title,
				Keyword_id:    p.Keyword_id,
				Created_at:    time.Now(),
			})
		}

		if len(productResult) > 0 {
			err = MongoTrainingData.StoreTrainingData(productResult)
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
