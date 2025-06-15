// db/mongo.go
package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func ConnectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}

	log.Println("✅ Connected to MongoDB!")

	MongoClient = client
}
