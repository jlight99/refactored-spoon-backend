package lib

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

const (
	dbConnStr = "mongodb+srv://RSdbuser:RSdbuserpw@cluster0-t4wa9.mongodb.net/test?retryWrites=true&w=majority"
)

func GetCollection(collectionName string) *mongo.Collection {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbConnStr))
	if err != nil {
		log.Fatalf("unable to create mongoDB client: %s\n", err.Error())
	}

	return client.Database("RefactoredSpoonDB").Collection(collectionName)
}
