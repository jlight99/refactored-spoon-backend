package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func MealsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	collection := lib.GetCollection("Days")
	userID := r.URL.Query().Get("user")

	switch r.Method {
	case http.MethodGet:
		getMeals(w, r, collection, userID, date)
	case http.MethodPost:
		postMeal(w, r, collection, userID, date)
	case http.MethodDelete:
		// deleteDay(w, r, collection, userID, date)
	}
}

func getMeals(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{"user": userID, "date": date})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not find day results:\n" + err.Error()))
		return
	}
	defer cur.Close(ctx)

	meals := make([]Meal, 0)

	for cur.Next(ctx) {
		var meal Meal
		err = cur.Decode(&meal)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		meals = append(meals, meal)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meals)
}

func postMeal(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	decoder := json.NewDecoder(r.Body)
	var meal Meal
	err := decoder.Decode(&meal)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode post meal request:\n" + err.Error()))
		return
	}
	if meal.ID == primitive.NilObjectID {
		meal.ID = primitive.NewObjectID()
	}
	for i, _ := range meal.Foods {
		if meal.Foods[i].ID == primitive.NilObjectID {
			meal.Foods[i].ID = primitive.NewObjectID()
		}
	}

	dayRecord := GetDayByDate(ctx, collection, userID, date)
	if dayRecord.ID == primitive.NilObjectID {
		dayRecord = &DayRecord{
			Date:  date,
			User:  userID,
			Meals: []Meal{},
			Nutrition: NutritionSummary{
				Calories: 0,
			},
		}
		_, err := collection.InsertOne(ctx, dayRecord)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to insert into day collection:\n" + err.Error()))
			return
		}
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"user": userID, "date": date},
		bson.M{
			"$set":  bson.M{"nutrition.calories": dayRecord.Nutrition.Calories + meal.Nutrition.Calories},
			"$push": bson.M{"meals": meal},
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to add meal into day collection:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	// w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
}

func deleteMeals(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string) {

}
