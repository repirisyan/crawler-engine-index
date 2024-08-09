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
	Title string    `bson:"title"`
	Link  string    `bson:"link"`
	Image *[]string `bson:"image"`
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
	Category    *string `bson:"category"`
	Location    struct {
		Country  *string
		Province *string
		City     *string
		District *string
	} `bson:"location"`
	Comodity struct {
		Comodity                  string
		Sub_comodity              *string
		Second_level_sub_comodity *string
		Third_level_sub_comodity  *string
	} `bson:"comodity"`
	Comodity_id    uint64  `bson:"comodity_id"`
	Keyword        string  `bson:"keyword"`
	Keyword_id     uint64  `bson:"keyword_id"`
	Marketplace    string  `bson:"marketplace"`
	Marketplace_id uint64  `bson:"marketplace_id"`
	User_id        uint64  `bson:"user_id"`
	Published_at   *string `bson:"published_at"`
	Created_at     string  `bson:"created_at"`
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
