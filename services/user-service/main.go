package main

import (
	"context"
	"encoding/json"
	"github.com/refactored-spoon-backend/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"time"
)

type userRequest struct {
	Username string
	Password string
}

func main() {
	log.Println("User Service start")

	collection := lib.GetCollection("Users")

	signupHandler := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var userReq userRequest
		err := decoder.Decode(&userReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("could not decode user signup request"))
			return
		}

		// verify on client side that both username and password are provided

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		res, err := collection.InsertOne(ctx, bson.M{"username": userReq.Username, "password": userReq.Password})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to insert into user collection"))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
	}

	loginHandler := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var userReq userRequest
		err := decoder.Decode(&userReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("could not decode user login request"))
			return
		}

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		res := collection.FindOne(ctx, bson.M{"username": userReq.Username, "password": userReq.Password})

		var findRes userRequest
		err = res.Decode(&findRes)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("could not find user with this login"))
			return
		}

		w.WriteHeader(http.StatusFound)
		w.Write([]byte(findRes.Username))
	}

	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login", loginHandler)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
