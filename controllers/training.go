package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
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

func CleanTrainingData() {
	fmt.Println("Clean Training Data")

	// Open the keywords JSON file
	keywordsFile, err := os.Open("assets/keywords.json")
	if err != nil {
		fmt.Printf("Error fetching keyword list: %v\n", err)
		return
	}
	defer keywordsFile.Close()

	// Decode the JSON file into a slice of keyword objects
	var keywords []struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(keywordsFile).Decode(&keywords); err != nil {
		fmt.Printf("Error decoding keywords JSON: %v\n", err)
		return
	}

	// Create a channel to collect updates
	updateChan := make(chan MongoTrainingData.TrainingData, 100) // Buffered channel

	// WaitGroup to synchronize goroutines
	var wg sync.WaitGroup

	// Variable to count the number of records changed
	var totalChangedRecords int64
	var mu sync.Mutex // Mutex to safely update the totalChangedRecords variable

	// Start a goroutine to collect and process updates
	go func() {
		var updateList []MongoTrainingData.TrainingData
		for update := range updateChan {
			updateList = append(updateList, update)
		}
		// Perform bulk update once all updates are collected
		if len(updateList) > 0 {
			if err := MongoTrainingData.UpdateTrainingData(updateList); err != nil {
				log.Printf("Failed to update training data: %v\n", err)
			} else {
				mu.Lock()
				totalChangedRecords += int64(len(updateList))
				mu.Unlock()
			}
		}
	}()

	// Process each keyword
	for _, kl := range keywords {
		wg.Add(1)
		go func(kl struct {
			ID   uint64 `json:"id"`
			Name string `json:"name"`
		}) {
			defer wg.Done()
			// Search for training data based on the keyword name
			trainingDatas, err := MongoTrainingData.SearchTrainingData(kl.Name)
			if err != nil {
				fmt.Printf("Error fetching Training Data for keyword '%s': %v\n", kl.Name, err)
				return
			}
			// Collect training data that needs to be updated
			for _, trainingData := range trainingDatas {
				// Update training data if the Keyword ID is different
				if trainingData.Keyword_id != kl.ID {
					updateChan <- MongoTrainingData.TrainingData{
						ID:            trainingData.ID,
						Keyword_id:    kl.ID,
						Product_title: trainingData.Product_title,
						Created_at:    time.Now(),
					}
				}
			}
		}(kl)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the channel to indicate that there are no more updates
	close(updateChan)

	// Print the total number of records changed
	fmt.Printf("Total records changed: %d\n", totalChangedRecords)
}
