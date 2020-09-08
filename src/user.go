package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/refactored-spoon-backend/src/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userRequest struct {
	ID       primitive.ObjectID `bson:"_id, omitempty"`
	Email    string
	Password string
}

func Signup(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Users")

	decoder := json.NewDecoder(r.Body)
	var userReq userRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode user signup request:\n" + err.Error()))
		return
	}

	// verify on client side that both email and password are provided

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, bson.M{"email": userReq.Email, "password": userReq.Password})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to insert into user collection:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
}

func Login(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Users")

	decoder := json.NewDecoder(r.Body)
	var userReq userRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode user login request:\n" + err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := collection.FindOne(ctx, bson.M{"email": userReq.Email, "password": userReq.Password})

	var findRes userRequest
	err = res.Decode(&findRes)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("could not find user with this login:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusFound)
	w.Write([]byte(findRes.ID.Hex()))
}
