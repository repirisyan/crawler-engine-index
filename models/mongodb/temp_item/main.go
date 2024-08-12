// models/mongodb/getTempItem.go
package MongoTempItem

import (
	"context"
	"crawler-index/db/mongodb"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	Title string
	Link  string
	Image *struct {
		Small *[]string
		Large *[]string
	}
	Price struct {
		Price          uint64
		Original_price uint64
		Discount       *uint32
	}
	Rating struct {
		Rating float64
		Count  uint64
	}
	Sold   uint64
	Seller struct {
		Name string
		Url  *string
	}
	Description *string
	Category    *string
	Location    struct {
		Country  *string
		Province *string
		City     *string
		District *string
	}
	Comodity struct {
		Comodity                  string
		Sub_comodity              *string
		Second_level_sub_comodity *string
		Third_level_sub_comodity  *string
	}
	Comodity_id    uint64
	Keyword        string
	Keyword_id     uint64
	Marketplace    string
	Marketplace_id uint64
	User_id        uint64
	Published_at   *string
	Created_at     string
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
	collection = client.Database(os.Getenv("DB_MONGO_DATABASE")).Collection("temp_items")
}

// GetAllProducts retrieves all products from the database
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

func SearchProduct(query string) ([]Product, error) {
	ctx := context.Background()

	filter := bson.M{
		"title": bson.M{
			"$regex": query,
		},
	}

	cursor, err := collection.Find(ctx, filter)
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

func DeleteCollection() {
	// Drop the collection
	err := collection.Drop(context.TODO())
	if err != nil {
		log.Fatalf("Failed to drop collection: %v", err)
	}

	fmt.Println("Collection dropped successfully")
}
