package MongoTrainingData

import (
	"context"
	"crawler-index/db/mongodb"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TrainingData struct {
	ID            primitive.ObjectID `bson:"_id"`
	Product_title string
	Keyword_id    uint64
	Created_at    time.Time
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

func SearchTrainingData(query string) ([]TrainingData, error) {
	ctx := context.Background()

	// Split the query into an array of words
	queryArray := strings.Split(query, " ")

	// Create an array of filters for each word in queryArray
	filters := []bson.M{}
	for _, word := range queryArray {
		// Add a regex filter for each word (case-insensitive)
		filters = append(filters, bson.M{
			"product_title": bson.M{
				"$regex":   word,
				"$options": "i", // 'i' for case-insensitive search
			},
		})
	}

	// Combine all filters using $and to require all words to be matched
	filter := bson.M{"$and": filters}

	// Perform the find operation
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice of TrainingData
	var trainingData []TrainingData
	for cursor.Next(ctx) {
		var product TrainingData
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		trainingData = append(trainingData, product)
	}

	// Check for any errors encountered during iteration
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return trainingData, nil
}

func UpdateTrainingData(updateList []TrainingData) error {
	var models []mongo.WriteModel

	for _, data := range updateList {
		// Create an update operation for each item in the updateList
		filter := bson.M{"_id": data.ID}
		update := bson.M{"$set": bson.M{"keyword_id": data.Keyword_id, "Created_at": data.Created_at}}

		// Append the update operation to the list of WriteModels
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}

	// Execute the bulk write operation
	bulkOption := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(context.TODO(), models, bulkOption)
	if err != nil {
		return err
	}

	return nil
}

func StoreTrainingData(training_data []interface{}) error {
	_, err := collection.InsertMany(context.TODO(), training_data)

	if err != nil {
		return fmt.Errorf("error inserting training_data: %v", err)
	}
	return nil
}
