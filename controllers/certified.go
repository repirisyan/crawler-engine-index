package controllers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"crawler-index/models/mongodb/temp_item" // Import the correct MongoDB package
)

type Certified struct {
	Bpom                bool
	Bpom_number         string
	Sni                 bool
	Halal               bool
	Distribution_permit bool
}

// Extracts and formats BPOM number from the description
func extractBPOM(description string) string {
	// Optimized pattern to match BPOM and POM codes with exactly two letters followed by a numeric or alphanumeric string
	pattern := `(?i)(?:BPOM|POM)\s*(?:No\.?\s*|:?\s*|RI\s*POM|RI\s*:|NA)?\s*([A-Z]{2})[-.\s]*([A-Z0-9]{7,})`

	// Compile regex
	re := regexp.MustCompile(pattern)

	// Find all matches in the description
	matches := re.FindAllStringSubmatch(description, -1)
	if len(matches) == 0 {
		return ""
	}

	for _, match := range matches {
		if len(match) > 2 {
			code := match[1]
			number := strings.ReplaceAll(match[2], " ", "")

			// Check if the number is numeric or valid alphanumeric based on the code
			if isValidBPOMNumber(code, number) {
				return code + number
			}
		}
	}
	return ""
}

// Helper function to validate BPOM number based on its code
func isValidBPOMNumber(code, number string) bool {
	// If the code is "NA" or "SD", the number can be alphanumeric
	if code == "NA" || code == "SD" {
		return true
	}

	// For other codes, the number must be numeric
	_, err := strconv.Atoi(number)
	return err == nil
}

// Function to set certified information for products in the MongoDB collection
func SetCertified() {
	fmt.Println("Setting Certifications...")
	const limit = 1000
	offset := 0

	for {
		// Fetch products in batches from the database
		products, err := MongoTempItem.GetAllProducts(offset, limit) // Correct the package name if needed
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}

		if len(products) == 0 {
			// No more records to fetch
			break
		}

		var productResult []interface{}
		for _, p := range products {
			if p.Description == nil {
				continue // Skip products without a description
			}

			bpomNumber := extractBPOM(*p.Description)

			// Append the product with its certification data
			productResult = append(productResult, MongoTempItem.Certified{
				ID: p.ID,
				Certified: Certified{
					Bpom:                strings.Contains(strings.ToLower(*p.Description), "bpom"),
					Bpom_number:         bpomNumber,
					Sni:                 strings.Contains(strings.ToLower(*p.Description), "sni"),
					Halal:               strings.Contains(strings.ToLower(*p.Description), "halal"),
					Distribution_permit: strings.Contains(strings.ToLower(*p.Description), "ijin edar"),
				},
			})
		}

		if len(productResult) > 0 {
			// Update products with certified information
			err = MongoTempItem.SetCertified(productResult)
			if err != nil {
				log.Printf("Failed to set certification: %v\n", err)
				continue
			}
		}

		// Update offset for the next batch
		offset += limit
	}
}
