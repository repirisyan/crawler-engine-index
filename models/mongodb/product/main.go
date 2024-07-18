// models/mongodb/getTempItem.go
package MongoProduct

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
	"crawler-index/db/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
)

type IndexProduct struct {
	Title       string  `bson:"title"`
	Link        string  `bson:"link"`
	Image       *string `bson:"image"`
	Price       uint64  `bson:"price"`
	Rating      float64 `bson:"rating"`
	Sold        uint64  `bson:"sold"`
	Seller      string  `bson:"seller"`
	Location    string  `bson:"location"`
	Comodity    string  `bson:"comodity"`
	Keyword     string  `bson:"keyword"`
	Marketplace string  `bson:"marketplace"`
	Created_at  string  `bson:"created_at"`
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

	uri := fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)
	fmt.Println(os.Getenv("DB_MONGO_PORT"))
	if err := mongodb.InitMongoDB(uri); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	// Get MongoDB client instance
	client := mongodb.GetMongoClient()
	if client == nil {
		log.Fatal("MongoDB client is nil")
	}

	// Set the collection
	collection = client.Database(os.Getenv("DB_MONGO_DATABASE")).Collection("products")
}


func StoreProducts(products []interface{}) error {
	_, err := collection.InsertMany(context.TODO(), products)

	if err != nil {
		return fmt.Errorf("error inserting products: %v", err)
	}
	return nil
}
