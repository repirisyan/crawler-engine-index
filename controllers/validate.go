package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"crawler-index/models/mongodb/temp_item"
	"crawler-index/models/mysql/temp_item"

	"github.com/joho/godotenv"
)

// Constants for reuse
const (
	contentType     = "application/json"
	maxProducts     = 1000
	sleepDuration   = 1 * time.Second
	envCategoryHost = "ML_CATEGORY_HOST"
)

// ValidateCategory validates the master category of products
func ValidateCategory() {
	fmt.Println("Validate Category")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
		return
	}

	offset := 0

	for {
		// Fetch products from MongoDB
		products, err := MongoTempItem.GetAllProducts(offset, maxProducts)
		if err != nil {
			log.Printf("Error fetching products: %v", err)
			break
		}

		var updateList []MongoTempItem.UpdateComodity

		// Iterate over fetched products
		for _, product := range products {
			if err := processProduct(product, &updateList); err != nil {
				log.Printf("Error processing product ID %v: %v", product.ID, err)
			}
		}

		// Perform the batch update if there are any items to update
		if len(updateList) > 0 {
			if err := MongoTempItem.UpdateProductComodity(updateList); err != nil {
				log.Printf("Error updating product commodities: %v", err)
			}
		}

		if len(products) < maxProducts {
			// No more records to fetch
			break
		}

		// Update offset for the next iteration
		offset += maxProducts

		// Optional: Add a sleep to avoid overloading the server
		time.Sleep(sleepDuration)
	}
}

// processProduct processes a single product, sends a POST request, and prepares updates if needed
func processProduct(product MongoTempItem.Product, updateList *[]MongoTempItem.UpdateComodity) error {
	data := map[string]interface{}{
		"product_title": product.Title,
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}

	// Make the POST request
	resp, err := http.Post(os.Getenv(envCategoryHost), contentType, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error making POST request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	var obj MongoTempItem.ValidateCategory
	if err := parseResponse(resp.Body, &obj); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	// Check if category needs to be updated
	if err := checkAndUpdateCategory(product, obj, updateList); err != nil {
		return fmt.Errorf("error checking/updating category: %w", err)
	}

	return nil
}

// parseResponse reads the response body and unmarshals it into the target struct
func parseResponse(body io.Reader, target interface{}) error {
	res, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	if err := json.Unmarshal(res, target); err != nil {
		return fmt.Errorf("error unmarshalling response: %w", err)
	}

	return nil
}

// checkAndUpdateCategory checks if the category needs updating and prepares the update list if needed
func checkAndUpdateCategory(product MongoTempItem.Product, obj MongoTempItem.ValidateCategory, updateList *[]MongoTempItem.UpdateComodity) error {
	keywordID, err := strconv.ParseUint(obj.Keyword_id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid keyword ID: %w", err)
	}

	if product.Keyword_id != keywordID && keywordID > 0 {
		// Get category details from MySQL
		category, err := mysqlTempItem.GetCategory(keywordID)
		if err != nil {
			return fmt.Errorf("error fetching category for keyword ID %d: %w", keywordID, err)
		}

		// Prepare UpdateComodity struct
		updateComodity := MongoTempItem.UpdateComodity{
			ID: product.ID,
			Comodity: MongoTempItem.Comodity{
				Comodity:                  category.Comodity,
				Sub_comodity:              category.SubComodity,
				Second_level_sub_comodity: category.SecondLevelSubComodity,
				Third_level_sub_comodity:  category.ThirdLevelSubComodity,
			},
			Keyword: category.Keyword,
		}

		// Append to the list of updates
		*updateList = append(*updateList, updateComodity)
	}

	return nil
}
