package main

import (
	"log"
	"net/http"

	"../../lib"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userRequest struct {
	ID       primitive.ObjectID `bson:"_id, omitempty"`
	Email    string
	Password string
}

func main() {
	log.Println("User Service start")

	collection := lib.GetCollection("Users")

	userService := NewUserService(collection)

	http.Handle("/signup", lib.CorsMiddleware(http.HandlerFunc(userService.Signup)))
	http.Handle("/login", lib.CorsMiddleware(http.HandlerFunc(userService.Login)))

	log.Fatal(http.ListenAndServe(":8081", nil))
}
