package main

import (
	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/lib"

	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	dateLayout = "2020-01-12"
)

type NutritionSummary struct {
	Calories int
}

type Food struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name,omitempty"`
	Group     string             `bson:"group,omitempty"`
	Serving   int                `bson:"serving,omitempty"`
	Nutrition NutritionSummary   `bson:"nutrition,omitempty"`
}

type Meal struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name,omitempty"`
	Foods     []Food             `bson:"foods,omitempty"`
	Nutrition NutritionSummary   `bson:"nutrition,omitempty"`
}

type DayRecord struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Date      string             `bson:"date,omitempty"`
	User      string             `bson:"user,omitempty"`
	Meals     []Meal             `bson:"meals,omitempty"`
	Nutrition NutritionSummary   `bson:"nutrition,omitempty"`
}

type getDayReq struct {
	User string
}

func main() {
	log.Println("Day Service start")

	handleRequests()
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.Handle("/days", lib.CorsMiddleware(http.HandlerFunc(DaysHandler))).Methods(http.MethodGet, http.MethodOptions)
	router.Handle("/days/{date}", lib.CorsMiddleware(http.HandlerFunc(DayHandler))).Methods(http.MethodGet, http.MethodDelete, http.MethodOptions)
	router.Handle("/days/{date}/meals", lib.CorsMiddleware(http.HandlerFunc(MealsHandler))).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions)
	router.Handle("/days/{date}/meals/{mealId}", lib.CorsMiddleware(http.HandlerFunc(MealHandler))).Methods(http.MethodGet, http.MethodDelete, http.MethodOptions)
	// router.Handle("/days/{date}/meals/{mealId}/foods", lib.CorsMiddleware(http.HandlerFunc(DaysHandler))).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	// router.Handle("/days/{date}/meals/{mealId}/foods/{foodId}", lib.CorsMiddleware(http.HandlerFunc(DaysHandler))).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	router.Handle("/day", lib.CorsMiddleware(http.HandlerFunc(day))).Methods(http.MethodOptions)

	router.Use(mux.CORSMethodMiddleware(router))

	log.Fatal(http.ListenAndServe(":8083", router))
}
