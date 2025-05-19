package controllers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	Crawler "crawler-index/models/postgres/crawler"
)

// Extracts and formats BPOM number from the description
func extractBPOM(description string) string {
	// Optimized pattern to match BPOM and POM codes with exactly two letters followed by a numeric or alphanumeric string
	pattern := `(?i)(?:BPOM|POM)?\s*(?:No\.?\s*|:?\s*|RI\s*POM\s*|RI\s*:|NA)?\s*([A-Z]{2})\s*[:\-.\s]*([A-Z0-9]{6,})`

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

			if containsAlphabetInFirstTwo(number) && isNumericAfterFirstTwo(number) {
				return number
			}

			// Check if the number is numeric or valid alphanumeric based on the code
			if isValidBPOMNumber(number) {
				return code + number
			}
		}
	}
	return ""
}

// Helper function to validate BPOM number based on its code
func isValidBPOMNumber(number string) bool {
	// For other codes, the number must be numeric
	_, err := strconv.Atoi(number)
	return err == nil
}

// Helper function to check if the first two characters contain an alphabet
func containsAlphabetInFirstTwo(number string) bool {
	if len(number) >= 2 {
		return isLetter(number[0]) || isLetter(number[1])
	}
	return false
}

// Helper function to check if the rest of the number (after the first two characters) is numeric
func isNumericAfterFirstTwo(number string) bool {
	if len(number) > 2 {
		rest := number[2:]
		_, err := strconv.Atoi(rest)
		return err == nil
	}
	return false
}

// Helper function to check if a character is a letter
func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// Function to set certified information for products in the MongoDB collection
func SetCertified() {
	fmt.Println("Setting Certifications...")
	const limit = 1000
	offset := 0
	// result := extractBPOM("DNA Salmon POM No: NA18230500673")
	// fmt.Println("Extracted BPOM Number:", result)
	for {
		// Fetch products in batches from the database
		products, err := Crawler.GetAllProductCertifications(offset, limit) // Correct the package name if needed
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}

		if len(products) == 0 {
			// No more records to fetch
			fmt.Printf("offset : %v", offset)
			break
		}

		var productResult []interface{}
		for _, p := range products {
			desc := strings.ToLower(p.Description) // Direct use of p.Description (no dereference needed)
			bpomNumber := extractBPOM(p.Description)

			// Append the product with its certification data
			productResult = append(productResult, Crawler.Certified{
				ID:                  p.ID,
				Bpom:                strings.Contains(desc, "bpom"),
				Bpom_number:         bpomNumber,
				Sni:                 strings.Contains(desc, "sni"),
				Halal:               strings.Contains(desc, "halal"),
				Distribution_permit: strings.Contains(desc, "ijin edar"),
			})
		}

		if len(productResult) > 0 {
			// Update products with certified information
			err = Crawler.SetCertified(productResult)
			if err != nil {
				log.Printf("Failed to set certification: %v\n", err)
				continue
			}
		}

		// Update offset for the next batch
		offset += limit
	}
}
