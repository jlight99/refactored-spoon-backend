package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/internal/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Counter struct {
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Count int    `json:"count,omitempty" bson:"count,omitempty"`
}

func CountersHandler(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("counter")

	switch r.Method {
	case http.MethodGet:
		getCounters(w, r, collection)
	case http.MethodPost:
		postCounter(w, r, collection)
	}
}

func CounterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection := lib.GetCollection("counter")

	switch r.Method {
	case http.MethodGet:
		getCounter(w, r, collection, name)
	case http.MethodPut:
		updateCounter(w, r, collection, name)
	}
}

func getCounters(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not find counter results:\n" + err.Error()))
		return
	}
	defer cur.Close(ctx)

	counters := make([]Counter, 0)

	for cur.Next(ctx) {
		var counter Counter
		err = cur.Decode(&counter)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		counters = append(counters, counter)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counters)
}

func getCounter(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := collection.FindOne(ctx, bson.M{"name": name})

	var counter Counter
	err := res.Decode(&counter)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("could not find counter with this name:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusFound)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counter)
}

func postCounter(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	decoder := json.NewDecoder(r.Body)
	var counter Counter
	err := decoder.Decode(&counter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode post counter request:\n" + err.Error()))
		return
	}

	res, err := collection.InsertOne(ctx, counter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to insert into counter collection:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
}

func updateCounter(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"name": name},
		bson.M{
			"$inc": bson.M{"count": 1},
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to increment counter:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
