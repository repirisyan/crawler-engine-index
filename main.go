package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"crawler-index/db/mongodb"
	"crawler-index/db/mysql"
	"crawler-index/models/mongodb/product"
	"crawler-index/models/mongodb/supervision"
	"crawler-index/models/mongodb/temp_item"
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

func main() {
	mysql.Init()
	defer mysql.DB.Close()
	removeDuplicateData("temp_items")
	// storeCrawlerData()
	storeSupervision()
	removeDuplicateData("supervisions")
	storeIndexData()
	removeDuplicateData("products")
	// storeMysqlSupervision()
	fmt.Printf("Cleaning Complete")
}

// Store Data from temp_item to products in mongodb
func storeIndexData() {
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

// Store Data from temp_item to supervisions in mongodb based on supervision list in mysql
func storeSupervision() {
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
					Title:          product.Title,
					Link:           product.Link,
					Image:          product.Image,
					Price:          product.Price,
					Sold:           product.Sold,
					Seller:         product.Seller,
					Description:    product.Description,
					Category:       product.Category,
					Location:       product.Location,
					Comodity:       product.Comodity,
					Comodity_id:    product.Comodity_id,
					Keyword:        product.Keyword,
					Keyword_id:     product.Keyword_id,
					Marketplace:    product.Marketplace,
					Marketplace_id: product.Marketplace_id,
					User_id:        product.User_id,
					Published_at:   product.Published_at,
					Status:         Status{Value: false},
					Crawler_at:     product.Created_at,
					Created_at:     formattedDate,
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
	err := godotenv.Load()
	// Initialize MongoDB client
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	uri := "mongodb://" + os.Getenv("DB_MONGO_HOST") + ":" + os.Getenv("DB_MONGO_PORT")
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
	keys := bson.D{
		{"title", 1},
		{"marketplace_id", 1},
		{"comodity_id", 1},
		{"created_at", 1},
		{"seller", 1},
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
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"title":          "$title",
					"marketplace_id": "$marketplace_id",
					"seller":         "$seller",
					"comodity_id":    "$comodity_id",
					"created_at":     "$created_at",
				},
				"duplicates": bson.M{"$addToSet": "$_id"},
				"count":      bson.M{"$sum": 1},
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gt": 1},
			},
		},
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
