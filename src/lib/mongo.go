package lib

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbConnStr = "mongodb+srv://ellen:creepy-elysium-just-like-THIS09@cluster0.wmwca.mongodb.net/refactored_spoon_db?retryWrites=true&w=majority"
)

func GetCollection(collectionName string) *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbConnStr))
	if err != nil {
		log.Fatalf("unable to create mongoDB client: %s\n", err.Error())
	}

	return client.Database("refactored_spoon_db").Collection(collectionName)
}
