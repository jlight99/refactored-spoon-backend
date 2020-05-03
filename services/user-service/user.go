package main

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type UserService struct {
	collection *mongo.Collection
}

func NewUserService(collection *mongo.Collection) *UserService {
	return &UserService{collection:collection}
}

func (u *UserService) Signup(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userReq userRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode user signup request"))
		return
	}

	// verify on client side that both email and password are provided

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := u.collection.InsertOne(ctx, bson.M{"email": userReq.Email, "password": userReq.Password})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to insert into user collection"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
}

func (u *UserService) Login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var userReq userRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode user login request"))
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res := u.collection.FindOne(ctx, bson.M{"email": userReq.Email, "password": userReq.Password})

	var findRes userRequest
	err = res.Decode(&findRes)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("could not find user with this login"))
		return
	}

	w.WriteHeader(http.StatusFound)
	w.Write([]byte(findRes.ID.Hex()))
}
