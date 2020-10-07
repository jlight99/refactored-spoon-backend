package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/internal/lib"
)

func main() {
	log.Println("refactored spoon server start")

	router := mux.NewRouter().StrictSlash(true)

	handleDayRequests(router)
	handleUserRequests(router)
	handleUSDARequests(router)

	// get port as environment variable since Heroku sets PORT variable dynamically
	// https://devcenter.heroku.com/articles/runtime-principles#web-servers
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8081"
	}

	server := &http.Server{
		Handler:      router,
		Addr:         ":" + port,
		WriteTimeout: 8 * time.Second,
		ReadTimeout:  8 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func handleDayRequests(router *mux.Router) {
	router.Handle("/days/{date}", lib.CorsMiddleware(http.HandlerFunc(DayHandler))).Methods(http.MethodGet, http.MethodDelete, http.MethodOptions)
	router.Handle("/days/{date}/meals", lib.CorsMiddleware(http.HandlerFunc(MealsHandler))).Methods(http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions)
	router.Handle("/days/{date}/meals/{mealId}", lib.CorsMiddleware(http.HandlerFunc(MealHandler))).Methods(http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodOptions)
}

func handleUserRequests(router *mux.Router) {
	router.Handle("/signup", lib.CorsMiddleware(http.HandlerFunc(Signup))).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/login", lib.CorsMiddleware(http.HandlerFunc(Login))).Methods(http.MethodPost, http.MethodOptions)
}

func handleUSDARequests(router *mux.Router) {
	router.Handle("/food/search", lib.CorsMiddleware(http.HandlerFunc(SearchFood))).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/food/detail", lib.CorsMiddleware(http.HandlerFunc(FoodDetail))).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/foods/detail", lib.CorsMiddleware(http.HandlerFunc(FoodsDetail))).Methods(http.MethodPost, http.MethodOptions)
}
