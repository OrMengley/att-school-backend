package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

func ConnectDB() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // default
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	fmt.Println("✅ MongoDB Connected")
	DB = client
}

func GetCollection(database string, collection string) *mongo.Collection {
	return DB.Database(database).Collection(collection)
}
