package lib

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbConnStr = os.Getenv("DB_CONN_STR")
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
