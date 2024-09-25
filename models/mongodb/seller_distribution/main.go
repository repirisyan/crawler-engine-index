// models/mongodb/getTempItem.go
package SellerDistribution

import (
	"context"
	"crawler-index/db/mongodb"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
)

type SellerDistribution struct {
	Comodity    string `bson:"comodity"`
	Marketplace string `bson:"marketplace"`
	Seller      int    `bson:"seller"`
	Year        int    `bson:"year"`
	Month       int    `bson:"month"`
	Date        string `bson:"date"`
}

var client *mongo.Client
var collection *mongo.Collection

func init() {
	// Connect to MongoDB
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongoHost := os.Getenv("DB_MONGO_HOST")
	mongoPort := os.Getenv("DB_MONGO_PORT")
	mongoUser := os.Getenv("DB_MONGO_USER")
	mongoAuthSource := os.Getenv("MONGO_AUTH_SOURCE")
	mongoPassword := os.Getenv("DB_MONGO_PASSWORD")

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=%s", mongoUser, mongoPassword, mongoHost, mongoPort, mongoAuthSource)

	if err := mongodb.InitMongoDB(uri); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	// Get MongoDB client instance
	client := mongodb.GetMongoClient()
	if client == nil {
		log.Fatal("MongoDB client is nil")
	}

	// Set the collection
	collection = client.Database(os.Getenv("DB_MONGO_DATABASE")).Collection("seller_distributions")
}

func StoreSellerDistribution(sellerDistribution []interface{}) error {
	_, err := collection.InsertMany(context.TODO(), sellerDistribution)

	if err != nil {
		return fmt.Errorf("error inserting products: %v", err)
	}
	return nil
}
