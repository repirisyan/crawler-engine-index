package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"crawler-index/db/mongodb"
	"crawler-index/db/mysql"
	"crawler-index/models/mongodb/product"
	"crawler-index/models/mongodb/supervision"
	"crawler-index/models/mongodb/temp_item"
	"crawler-index/models/mysql/supervision"
	"crawler-index/models/mysql/supervision_list"
	"crawler-index/models/mysql/temp_item"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mysql.Init()
	defer mysql.DB.Close()
	removeDuplicateData("temp_items")
	storeCrawlerData()
	storeSupervision()
	removeDuplicateData("supervisions")
	storeIndexData()
	removeDuplicateData("products")
	storeMysqlSupervision()
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
				Category: 	 p.Category,
				Link:        p.Link,
				Image:       p.Image,
				Price:       p.Price,
				Rating:      p.Rating,
				Sold:        p.Sold,
				Seller:      p.Seller,
				Location:    p.Location,
				Comodity:    p.Comodity,
				Sub_comodity: p.Sub_comodity,
				Second_level_sub_comodity: p.Second_level_sub_comodity,
				Third_level_sub_comodity: p.Third_level_sub_comodity,
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

// Store Data from temp_item to supervisions in mongodb based on search key
func storeSupervision(){
	for {
		supervisionList, err := mysqlSupervisionList.GetAllData()
		if err != nil {
			fmt.Printf("Error fetching supervision List: %v\n", err)
			break
		}

		for _, svl := range supervisionList {
			var productResult []interface{}
			temp_items, err := mysqlTempItem.GetAllData(svl.Name)
			if err != nil {
				fmt.Printf("Error fetching Temp Item: %v\n", err)
				break
			}
			for _, temp_item := range temp_items {
				productResult = append(productResult, MongoSupervision.Product{
					Title:       temp_item.Title,
					Link:        temp_item.Link,
					Image:       temp_item.Image,
					Price:       temp_item.Price,
					Sold:        temp_item.Sold,
					Seller:      temp_item.Seller,
					Location:    temp_item.Location,
					Keyword_id:    temp_item.Keyword_id,
					Marketplace_id: temp_item.Marketplace_id,
					Supervision_id:  temp_item.Supervision_id,
					Created_at: temp_item.Created_at,
				})
				if err != nil {
					log.Printf("Failed to insert supervision: %v\n", err)
					continue
				}
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

// Optimized
func storeMysqlSupervision() {
    limit := 1000
    offset := 0

    for {
        products, err := MongoSupervision.GetAllProducts(offset, limit)
        if err != nil {
            fmt.Printf("Error fetching Mongo Supervision: %v\n", err)
            break
        }

        if len(products) == 0 {
            // No more records to fetch
            break
        }

        var supervisions []mysqlSupervision.Supervision
        var updateFlags []mysqlTempItem.FlagUpdate

        for _, product := range products {
            value := mysqlSupervision.Supervision{
                Name:           product.Title,
                Link:           product.Link,
                Image:          product.Image,
                Price:          product.Price,
                Sold:           product.Sold,
                Seller:         product.Seller,
                Location:       product.Location,
                Keyword_id:     product.Keyword_id,
                Marketplace_id: product.Marketplace_id,
                Created_at:     product.Created_at,
            }
            supervisions = append(supervisions, value)

            flagUpdate := mysqlTempItem.FlagUpdate{
                Title:          product.Title,
                Seller:         product.Seller,
                Marketplace_id: product.Marketplace_id,
            }
            updateFlags = append(updateFlags, flagUpdate)
        }

        if len(supervisions) > 0 {
            err = mysqlSupervision.StoreSupervisions(supervisions)
            if err != nil {
                log.Printf("Failed to insert batch products: %v\n", err)
                continue
            }
        }

        if len(updateFlags) > 0 {
            err = mysqlTempItem.UpdateFlags(updateFlags)
            if err != nil {
                log.Printf("Failed to update product flags: %v\n", err)
                continue
            }
        }

        // Update offset for next iteration
        offset += limit
    }

    // Optionally delete collections if necessary
    MongoSupervision.DeleteCollection()
    MongoTempItem.DeleteCollection()
}


// Store Data from temp_item from mongodb to temp_item in mysql
// Optimized
func storeCrawlerData() {
    limit := 1000
    offset := 0

    for {
        products, err := MongoTempItem.GetAllProducts(offset, limit)
        if err != nil {
            fmt.Printf("Error fetching Mongo products: %v\n", err)
            break
        }

        if len(products) == 0 {
            // No more records to fetch
            break
        }

        var batch []mysqlTempItem.Product
        for _, product := range products {
            value := mysqlTempItem.Product{
                Title:          product.Title,
                Link:           product.Link,
                Image:          product.Image,
                Price:          product.Price,
                Rating:         product.Rating,
                Sold:           product.Sold,
                Seller:         product.Seller,
                Location:       product.Location,
                Keyword_id:     product.Keyword_id,
                Marketplace_id: product.Marketplace_id,
                User_id:        product.User_id,
                Created_at:     product.Created_at,
            }
            batch = append(batch, value)
        }

        if len(batch) > 0 {
            err = mysqlTempItem.StoreProducts(batch)  // Assuming StoreProducts handles batch insert
            if err != nil {
                log.Printf("Failed to insert batch: %v\n", err)
                continue
            }
        }

        // Update offset for next iteration
        offset += limit
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
	indexOptions := options.Index().SetBackground(true).SetUnique(false)
	keys := bson.D{
		{"title", 1},
		{"marketplace_id", 1},
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
					"title":       "$title",
					"marketplace_id": "$marketplace_id",
					"seller":      "$seller",
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
