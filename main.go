package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"strconv"

	"crawler-index/db/mongodb"
	"crawler-index/db/mysql"
	"crawler-index/models/mongodb/product"
	"crawler-index/models/mongodb/supervision"
	"crawler-index/models/mongodb/temp_item"
	"crawler-index/models/mongodb/train_data"
	"crawler-index/models/mysql/supervision_list"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Status struct {
	Value bool
}

type Certified struct {
	Bpom                bool
	Bpom_number         string
	Sni                 bool
	Halal               bool
	Distribution_permit bool
}

func main() {
	mysql.Init()
	defer mysql.DB.Close()
	// removeDuplicateData("temp_items")
	setCertified()

	// validateCategory()
	// storeTrainingData()
	// removeDuplicateData("training_data")
	// storeSupervision()
	// removeDuplicateData("supervisions")
	// storeIndexData()
	// removeDuplicateData("products")
	fmt.Printf("Cleaning Complete")
}

// Extracts and formats BPOM number from the description
func extractBPOM(description string) string {
    // Updated pattern to match BPOM and POM codes with exactly two letters followed by a numeric string
    pattern := `(?i)(?:BPOM|POM)\s*(?:No\.?\s*|:?\s*|RI\s*POM|RI\s*:)?\s*([A-Z]{2})[-.\s]*(\d{7,}|[A-Z0-9]{9,})`

    // Compile regex
    re := regexp.MustCompile(pattern)

    // Find matches
    matches := re.FindAllStringSubmatch(description, -1)
    if len(matches) > 0 {
        for _, match := range matches {
            if len(match) > 2 {
                code := match[1]
                number := strings.ReplaceAll(match[2], " ", "")

                // Ensure the number is numeric and the code is exactly two letters
                if _, err := strconv.Atoi(number); err == nil || code == "NA" || code == "SD" {
                    return code + number
                }
            }
        }
    }
    return ""
}

func setCertified() {
	fmt.Println("Set Certificated")
	limit := 1000
	offset := 0
	for {
		products, err := MongoTempItem.GetAllProducts(offset, limit)
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}

		var productResult []interface{}
		for _, p := range products {
			bpom_number := extractBPOM(*p.Description)

			productResult = append(productResult, MongoTempItem.Certified{
				ID: p.ID,
				Certified: Certified{
					Bpom:                strings.Contains(*p.Description, "BPOM"),
					Bpom_number:         bpom_number,
					Sni:                 strings.Contains(*p.Description, "SNI"),
					Halal:               strings.Contains(*p.Description, "Halal"),
					Distribution_permit: strings.Contains(*p.Description, "Ijin Edar"),
				},
			})
		}

		if len(productResult) > 0 {
			err = MongoTempItem.SetCertified(productResult)
			if err != nil {
				log.Printf("Failed to set supervision category: %v\n", err)
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

// Store Data from temp_item to products in mongodb
func storeIndexData() {
	fmt.Println("Store Index Data")

	limit := 1000
	offset := 0
	for {
		products, err := MongoTempItem.GetAllProducts(offset, limit)
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}
		// Convert struct to interface
		var productResult []interface{}
		for _, p := range products {
			productResult = append(productResult, MongoProduct.IndexProduct{
				Title:       p.Title,
				Description: p.Description,
				Category:    p.Category,
				Link:        p.Link,
				Image:       p.Image,
				Price:       p.Price,
				Rating:      p.Rating,
				Sold:        p.Sold,
				Seller:      p.Seller,
				Location:    p.Location,
				Certified:   p.Certified,
				Comodity:    p.Comodity,
				Keyword:     p.Keyword,
				Marketplace: p.Marketplace,
				Created_at:  p.Created_at,
			})
		}

		if len(productResult) > 0 {
			err = MongoProduct.StoreProducts(productResult)
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

// Validate master category
func validateCategory() {
	fmt.Println("Validate Category")

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	limit := 1000
	offset := 0

	for {
		products, err := MongoTempItem.GetDataForTraining(offset, limit)
		if err != nil {
			log.Printf("Error fetching products: %v\n", err)
			break
		}

		var updateList []interface{}
		for _, p := range products {
			data := map[string]string{
				"product_title":    p.Title,
				"crawler_category": p.Category,
			}

			// Convert data to JSON
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Printf("Error marshalling data: %v\n", err)
				return
			}
			fmt.Println("Data:", data)
			// Make the POST request
			resp, err := http.Post(os.Getenv("ML_CATEGORY_HOST"), "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error making POST request: %v\n", err)
				continue // Continue with the next product
			}
			defer resp.Body.Close()

			// Read the response
			res, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response: %v\n", err)
				continue // Continue with the next product
			}

			var obj MongoTrainingData.TrainingData

			err = json.Unmarshal(res, &obj)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			// Convert response to string

			if obj.Master_category != p.Comodity.Comodity {
				updateList = append(updateList, MongoTempItem.UpdateComodity{
					ID: p.ID,
					Comodity: struct {
						Comodity                  string
						Sub_comodity              *string
						Second_level_sub_comodity *string
						Third_level_sub_comodity  *string
					}{
						obj.Master_category,
						obj.Sub_master_category,
						obj.Second_level_sub_master_category,
						obj.Third_level_sub_master_category,
					},
					Keyword: obj.Keyword,
				})
			}
		}

		if len(updateList) > 0 {
			err = MongoTempItem.UpdateProductComodity(updateList)
			if err != nil {
				log.Printf("Failed to perform batch update: %v\n", err)
				// Continue to next batch even if update fails
			}
		}

		if len(products) < limit {
			// No more records to fetch
			break
		}

		// Update offset for next iteration
		offset += limit

		// Optional: Add a sleep to avoid overloading the server
		time.Sleep(1 * time.Second)
	}
}

// Training Data for Machine Learning label Category
func storeTrainingData() {
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
				Product_title:                    p.Title,
				Crawler_category:                 p.Category,
				Master_category:                  p.Comodity.Comodity,
				Sub_master_category:              p.Comodity.Sub_comodity,
				Second_level_sub_master_category: p.Comodity.Second_level_sub_comodity,
				Third_level_sub_master_category:  p.Comodity.Third_level_sub_comodity,
				Keyword:                          p.Keyword,
				Created_at:                       time.Now(),
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

// Store Data from temp_item to supervisions in mongodb based on supervision list in mysql
func storeSupervision() {
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

// Remove Duplicate Data from products in mongodb
func removeDuplicateData(collection_name string) {
	fmt.Println("Remove Duplicate Data From %s", collection_name)

	err := godotenv.Load()
	// Initialize MongoDB client
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	uri := "mongodb://" + os.Getenv("DB_MONGO_USER") + ":" + os.Getenv("DB_MONGO_PASSWORD") + "@" + os.Getenv("DB_MONGO_HOST") + ":" + os.Getenv("DB_MONGO_PORT")
	if err := mongodb.InitMongoDB(uri); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	// Get MongoDB client instance
	client := mongodb.GetMongoClient()
	if client == nil {
		log.Fatal("MongoDB client is nil")
	}

	// Specify the database and collection
	collectionName := collection_name

	// Access the collection
	collection := client.Database(os.Getenv("DB_MONGO_DATABASE")).Collection(collectionName)

	// Create an index on the fields to identify duplicates
	indexOptions := options.Index().SetUnique(false)
	var keys bson.D

	switch collectionName {
	case "training_data":
		keys = bson.D{{"product_title", 1}, {"crawler_category", 1}, {"keyword", 1}}
	case "supervisions":
		keys = bson.D{{"title", 1}, {"marketplace", 1}, {"supervision_category", 1}, {"seller", 1}}
	default:
		keys = bson.D{{"title", 1}, {"marketplace", 1}, {"seller", 1}}
	}

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: indexOptions,
	}

	// Create the index on the collection
	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	// Aggregation pipeline to identify and delete duplicates
	var pipeline []bson.M
	switch collectionName {
	case "training_data":
		pipeline = []bson.M{
			{"$group": bson.M{
				"_id":        bson.M{"product_title": "$product_title"},
				"duplicates": bson.M{"$addToSet": "$_id"},
				"count":      bson.M{"$sum": 1},
			}},
			{"$match": bson.M{"count": bson.M{"$gt": 1}}},
		}
	case "supervisions":
		pipeline = []bson.M{
			{"$group": bson.M{
				"_id":        bson.M{"title": "$title", "marketplace": "$marketplace", "supervision_category": "$supervision_category", "seller": "$seller"},
				"duplicates": bson.M{"$addToSet": "$_id"},
				"count":      bson.M{"$sum": 1},
			}},
			{"$match": bson.M{"count": bson.M{"$gt": 1}}},
		}
	default:
		pipeline = []bson.M{
			{"$group": bson.M{
				"_id":        bson.M{"title": "$title", "marketplace": "$marketplace", "seller": "$seller"},
				"duplicates": bson.M{"$addToSet": "$_id"},
				"count":      bson.M{"$sum": 1},
			}},
			{"$match": bson.M{"count": bson.M{"$gt": 1}}},
		}
	}

	// Perform aggregation to identify duplicates
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatalf("Aggregation error: %v", err)
	}
	defer cursor.Close(context.Background())

	// Iterate over the cursor to delete duplicates
	for cursor.Next(context.Background()) {
		var result struct {
			ID         bson.M        `bson:"_id"`
			Duplicates []interface{} `bson:"duplicates"`
			Count      int           `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			log.Printf("Error decoding result: %v", err)
			continue
		}

		// Get duplicates to delete
		duplicates := result.Duplicates

		var wg sync.WaitGroup
		for i := 1; i < len(duplicates); i++ {
			wg.Add(1)
			go func(id interface{}) {
				defer wg.Done()
				idToDelete, ok := id.(primitive.ObjectID)
				if !ok {
					log.Printf("Failed to convert to ObjectID: %+v", id)
					return
				}
				_, err := collection.DeleteOne(context.Background(), bson.M{"_id": idToDelete})
				if err != nil {
					log.Printf("Failed to delete duplicate: %v", err)
				}
			}(duplicates[i])
		}
		wg.Wait()
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
	}
}
