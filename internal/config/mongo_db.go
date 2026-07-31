package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var DB *mongo.Database

func ConnectMongo(uri, dbName string) *mongo.Database {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))

	if err != nil {
		log.Fatal("Error connecting to DB")
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Mongo Not responding:", err)
	}

	return client.Database(dbName)

}
