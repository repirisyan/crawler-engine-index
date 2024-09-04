package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"crawler-index/db/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/joho/godotenv"
)

func RemoveDuplicationData(collectionName string) {
	fmt.Printf("Removing duplicate data from %s \n", collectionName)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize MongoDB client
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", os.Getenv("DB_MONGO_USER"), os.Getenv("DB_MONGO_PASSWORD"), os.Getenv("DB_MONGO_HOST"), os.Getenv("DB_MONGO_PORT"))
	if err := mongodb.InitMongoDB(uri); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	// Get MongoDB client instance
	client := mongodb.GetMongoClient()
	if client == nil {
		log.Fatal("MongoDB client is nil")
	}

	// Access the collection
	collection := client.Database(os.Getenv("DB_MONGO_DATABASE")).Collection(collectionName)

	// Create an index on the fields to identify duplicates
	indexModel := createIndexModel(collectionName)
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	// Build aggregation pipeline to identify duplicates
	pipeline := buildAggregationPipeline(collectionName)

	// Perform aggregation to identify duplicates
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatalf("Aggregation error: %v", err)
	}
	defer cursor.Close(context.Background())

	// Iterate over the cursor to delete duplicates
	var wg sync.WaitGroup
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

		// Remove duplicates, keeping one entry
		if len(result.Duplicates) > 1 {
			removeDuplicates(collection, result.Duplicates[1:], &wg)
		}
	}
	wg.Wait()

	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
	}
	fmt.Printf("Duplicate removal completed for %s.\n", collectionName)
}

func createIndexModel(collectionName string) mongo.IndexModel {
	var keys bson.D

	switch collectionName {
	case "training_data":
		keys = bson.D{{Key: "product_title", Value: 1}, {Key: "keyword_id", Value: 1}}
	case "supervisions":
		keys = bson.D{{Key: "title", Value: 1}, {Key: "marketplace", Value: 1}, {Key: "supervision_category", Value: 1}, {Key: "seller", Value: 1}}
	default:
		keys = bson.D{{Key: "title", Value: 1}, {Key: "marketplace", Value: 1}, {Key: "seller", Value: 1}}
	}

	indexOptions := options.Index().SetUnique(false)
	return mongo.IndexModel{
		Keys:    keys,
		Options: indexOptions,
	}
}

func buildAggregationPipeline(collectionName string) []bson.M {
	var pipeline []bson.M

	switch collectionName {
	case "training_data":
		pipeline = []bson.M{
			{"$group": bson.M{
				"_id":        bson.M{"product_title": "$product_title", "keyword_id": "$keyword_id"},
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

	return pipeline
}

func removeDuplicates(collection *mongo.Collection, duplicates []interface{}, wg *sync.WaitGroup) {
	for _, id := range duplicates {
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
		}(id)
	}
}
