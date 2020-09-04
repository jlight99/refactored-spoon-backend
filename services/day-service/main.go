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
	Calories int `json:"calories,omitempty" bson:"calories,omitempty"`
}

type Food struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty"`
	Group     string             `json:"group,omitempty" bson:"group,omitempty"`
	Serving   int                `json:"serving,omitempty" bson:"serving,omitempty"`
	Nutrition NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
}

type Meal struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty"`
	Foods     []Food             `json:"foods,omitempty" bson:"foods,omitempty"`
	Nutrition NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
}

type DayRecord struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Date      string             `json:"date,omitempty" bson:"date,omitempty"`
	User      string             `json:"user,omitempty" bson:"user,omitempty"`
	Meals     []Meal             `json:"meals,omitempty" bson:"meals,omitempty"`
	Nutrition NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
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
	router.Handle("/days/{date}/meals", lib.CorsMiddleware(http.HandlerFunc(MealsHandler))).Methods(http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions)
	router.Handle("/days/{date}/meals/{mealId}", lib.CorsMiddleware(http.HandlerFunc(MealHandler))).Methods(http.MethodGet, http.MethodDelete, http.MethodPut, http.MethodOptions)

	router.Use(mux.CORSMethodMiddleware(router))

	log.Fatal(http.ListenAndServe(":8083", router))
}
