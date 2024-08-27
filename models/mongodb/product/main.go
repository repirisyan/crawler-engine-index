// models/mongodb/getTempItem.go
package MongoProduct

import (
	"context"
	"crawler-index/db/mongodb"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
)

type IndexProduct struct {
	Title string `bson:"title"`
	Link  string `bson:"link"`
	Image *struct {
		Small *[]string
		Large *[]string
	} `bson:"image"`
	Price struct {
		Price          uint64
		Original_price uint64
		Discount       *uint32
	} `bson:"price"`
	Rating struct {
		Rating float64
		Count  uint64
	} `bson:"rating"`
	Sold   uint64 `bson:"sold"`
	Seller struct {
		Name string
		Url  *string
	} `bson:"seller"`
	Description *string `bson:"description"`
	Category    string  `bson:"category"`
	Location    struct {
		Country  *string
		Province *string
		City     *string
		District *string
	} `bson:"location"`
	Certified struct {
		Bpom                bool
		Sni                 bool
		Halal               bool
		Distribution_permit bool
	}
	Comodity struct {
		Comodity                  string
		Sub_comodity              *string
		Second_level_sub_comodity *string
		Third_level_sub_comodity  *string
	} `bson:"comodity"`
	Keyword      string  `bson:"keyword"`
	Marketplace  string  `bson:"marketplace"`
	Published_at *string `bson:"published_at"`
	Created_at   string  `bson:"created_at"`
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
	mongoPassword := os.Getenv("DB_MONGO_PASSWORD")

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", mongoUser, mongoPassword, mongoHost, mongoPort)
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
