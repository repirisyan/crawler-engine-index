package main

import (
	"context"
	"crawler-index/db/mongo"
	"crawler-index/db/mysql"
	"crawler-index/models/mysql"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func main() {
	insertData()
	removeDuplicateData()
}

func insertData() {
	mysql.Init()
	if err := mongodb.InitMongoDB("mongodb://localhost:27017"); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	limit := 1000
	offset := 0

	client := mongodb.GetMongoClient()
	if client == nil {
		log.Fatal("MongoDB client is nil")
	}

	collection := client.Database("crawler").Collection("products")

	for {
		products, err := product.GetAllProduct(offset, limit)
		if err != nil {
			fmt.Printf("Error fetching products: %v\n", err)
			break
		}
		for _, product := range products {
			// Create a context for the operation
			ctx := context.Background()

			bsonProduct, err := bson.Marshal(product)
			if err != nil {
				log.Printf("Failed to marshal product: %v\n", err)
				continue
			}
			_, err = collection.InsertOne(ctx, bsonProduct)
			if err != nil {
				log.Printf("Failed to insert product %s: %v\n", product.Title, err)
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
	defer mysql.DB.Close()

	defer mongodb.CloseMongoDB()
}

func removeDuplicateData() {
	// Initialize MongoDB client
	if err := mongodb.InitMongoDB("mongodb://localhost:27017"); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer mongodb.CloseMongoDB()

	// Get MongoDB client instance
	client := mongodb.GetMongoClient()
	if client == nil {
		log.Fatal("MongoDB client is nil")
	}

	// Specify the database and collection
	databaseName := "crawler"
	collectionName := "products"

	// Access the collection
	collection := client.Database(databaseName).Collection(collectionName)

	// Create an index on the fields to identify duplicates
	indexOptions := options.Index().SetBackground(true).SetUnique(false)
	keys := bson.D{
		{"title", 1},
		{"marketplace", 1},
		{"seller", 1},
	}
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: indexOptions,
	}

	// Create the index on the collection
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	// Aggregation pipeline to identify and delete duplicates
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"title":       "$title",
					"marketplace": "$marketplace",
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

		// Remove all duplicates except one
		for i := 1; i < len(duplicates); i++ {
			idToDelete, ok := duplicates[i].(primitive.ObjectID) // Assuming _id is ObjectID, adjust as per your schema
			if !ok {
				log.Printf("Failed to convert to ObjectID: %+v", duplicates[i])
				continue
			}

			_, err := collection.DeleteOne(context.Background(), bson.M{"_id": idToDelete})
			if err != nil {
				log.Printf("Failed to delete duplicate: %v", err)
				continue
			}
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
	}

	log.Println("Duplicate removal process completed")
}
