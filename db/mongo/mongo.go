package mongodb

import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
)

type MongoDB struct {
    client *mongo.Client
    ctx    context.Context
}

// Global MongoDB instance
var db *MongoDB

// InitMongoDB initializes the MongoDB client
func InitMongoDB(uri string) error {
    clientOptions := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return err
    }

    // Check the connection
    err = client.Ping(context.Background(), nil)
    if err != nil {
        return err
    }

    db = &MongoDB{
        client: client,
        ctx:    context.Background(),
    }

    return nil
}

// CloseMongoDB closes the MongoDB client connection
func CloseMongoDB() {
    if db != nil && db.client != nil {
        if err := db.client.Disconnect(db.ctx); err != nil {
            log.Fatal(err)
        }
    }
}

// GetMongoClient returns the MongoDB client instance
func GetMongoClient() *mongo.Client {
    return db.client
}
