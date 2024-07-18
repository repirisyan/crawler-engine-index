// models/mongodb/getTempItem.go
package MongoSupervision

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/joho/godotenv"
	"crawler-index/db/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	Title       string  `bson:"title"`
	Link        string  `bson:"link"`
	Image       *string `bson:"image"`
	Price       uint64  `bson:"price"`
	Sold        uint64  `bson:"sold"`
	Seller      string  `bson:"seller"`
	Location    string  `bson:"location"`
	Comodity_id uint64  `bson:"comodity_id"`
	Marketplace_id uint64  `bson:"marketplace_id"`
	Supervision_id  uint64  `bson:"id"`
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
	collection = client.Database(os.Getenv("DB_MONGO_DATABASE")).Collection("supervisions")
}


func StoreProducts(products []interface{}) error {
	_, err := collection.InsertMany(context.TODO(), products)

	if err != nil {
		return fmt.Errorf("error inserting products: %v", err)
	}
	return nil
}

func GetAllProducts(offset, limit int) ([]Product, error) {
	ctx := context.Background()

	// Define options for find
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []Product
	for cursor.Next(ctx) {
		var product Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func DeleteCollection(){
	// Drop the collection
    err := collection.Drop(context.TODO())
    if err != nil {
        log.Fatalf("Failed to drop collection: %v", err)
    }
}
