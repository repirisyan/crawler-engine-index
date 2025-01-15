package controllers

import (
	"fmt"
	"log"
	"time"

	"crawler-index/models/mongodb/supervision"
	"crawler-index/models/mongodb/temp_item"
	"crawler-index/models/mysql/supervision_list"
)

type Status struct {
	Value bool
}

// Store Data from temp_item to supervisions in mongodb based on supervision list in mysql
func StoreSupervision() {
	fmt.Println("Store Supervision")

	for {
		supervisionList, err := mysqlSupervisionList.GetAllData()
		if err != nil {
			fmt.Printf("Error fetching supervision List: %v\n", err)
			break
		}
		now := time.Now()
		formattedDate := now.Format("2006-01-02")
		for _, svl := range supervisionList {
			var productResult []interface{}
			products, err := MongoTempItem.SearchProduct(svl.Name)
			if err != nil {
				fmt.Printf("Error fetching Temp Item: %v\n", err)
				break
			}
			for _, product := range products {
				productResult = append(productResult, MongoSupervision.Product{
					Title:               product.Title,
					Link:                product.Link,
					Image:               product.Image,
					Price:               product.Price,
					Sold:                product.Sold,
					Seller:              product.Seller,
					Description:         product.Description,
					Category:            product.Category,
					Location:            product.Location,
					Comodity:            product.Comodity,
					Keyword:             product.Keyword,
					Certified:           product.Certified,
					Supervision_keyword: svl.Name,
					Supervision_label:   svl.Category,
					Marketplace:         product.Marketplace,
					Published_at:        product.Published_at,
					Status:              Status{Value: false},
					Crawler_at:          product.Created_at,
					Created_at:          formattedDate,
				})
			}
			if len(productResult) > 0 {
				err = MongoSupervision.StoreProducts(productResult)
				if err != nil {
					log.Printf("Failed to insert batch products: %v\n", err)
					continue
				}
			}
		}
		// No more records to fetch
		break
	}
}
