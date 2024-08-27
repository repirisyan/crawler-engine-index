package MongoTrainingData

import (
	"context"
	"crawler-index/db/mongodb"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
)

type TrainingData struct {
	Product_title                    string
	Crawler_category                 string
	Master_category                  string
	Sub_master_category              *string
	Second_level_sub_master_category *string
	Third_level_sub_master_category  *string
	Keyword                          string
	Created_at                       time.Time
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

	if err := mongodb.InitMongoDB(uri); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}

	// Get MongoDB client instance
	client := mongodb.GetMongoClient()
	if client == nil {
		log.Fatal("MongoDB client is nil")
	}

	// Set the collection
	collection = client.Database(os.Getenv("DB_MONGO_DATABASE")).Collection("training_data")
}

func StoreTrainingData(training_data []interface{}) error {
	_, err := collection.InsertMany(context.TODO(), training_data)

	if err != nil {
		return fmt.Errorf("error inserting training_data: %v", err)
	}
	return nil
}
