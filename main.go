package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/internal/lib"
)

func main() {
	log.Println("refactored spoon server start")

	router := mux.NewRouter().StrictSlash(true)

	handleDayRequests(router)
	handleUserRequests(router)
	handleUSDARequests(router)

	router.Use(mux.CORSMethodMiddleware(router))

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8081"
	}

	log.Fatal(http.ListenAndServe(":"+port, router))
	log.Println("refactored spoon server start on port " + port)
}

func handleDayRequests(router *mux.Router) {
	router.Handle("/days", lib.CorsMiddleware(http.HandlerFunc(DaysHandler))).Methods(http.MethodGet, http.MethodOptions)
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
