package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	}
}

func MealHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]
	mealID := vars["mealId"]
	mealObjectID, err := primitive.ObjectIDFromHex(mealID)
	if err != nil {
		log.Println("error: invalid meal ID provided: " + mealID)
		return
	}

	collection := lib.GetCollection("Days")
	userID := r.URL.Query().Get("user")

	switch r.Method {
	case http.MethodDelete:
		deleteMeal(w, r, collection, userID, date, mealObjectID)
	case http.MethodPut:
		updateMeal(w, r, collection, userID, date, mealObjectID)
	}
}

func updateMeal(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string, mealID primitive.ObjectID) {
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

	for i, _ := range meal.Foods {
		if meal.Foods[i].ID == primitive.NilObjectID {
			meal.Foods[i].ID = primitive.NewObjectID()
		}
	}

	dayRecord := GetDayByDate(ctx, collection, userID, date)
	meals := dayRecord.Meals
	var deleteIdx int
	for i, _ := range meals {
		if meals[i].ID == mealID {
			deleteIdx = i
			break
		}
	}

	deleteCals := meals[deleteIdx].Nutrition.Calories

	meals[deleteIdx] = meal

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"user": userID, "date": date},
		bson.M{
			"$set": bson.M{"nutrition.calories": dayRecord.Nutrition.Calories - deleteCals + meal.Nutrition.Calories, "meals": meals},
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to add meal into day collection:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteMeal(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string, mealID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dayRecord := GetDayByDate(ctx, collection, userID, date)
	var mealToDelete Meal
	for _, meal := range dayRecord.Meals {
		if meal.ID == mealID {
			mealToDelete = meal
			break
		}
	}
	collection.UpdateOne(
		ctx,
		bson.M{"user": userID, "date": date, "meals": bson.M{"$elemMatch": bson.M{"_id": mealID}}},
		bson.M{
			"$set":  bson.M{"nutrition.calories": dayRecord.Nutrition.Calories - mealToDelete.Nutrition.Calories},
			"$pull": bson.M{"meals": mealToDelete},
		},
	)
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
			fmt.Println("insert dayRecord failed!")
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to insert into day collection:\n" + err.Error()))
			return
		}
	}

	nutrition := dayRecord.Nutrition
	nutrition.Calories += meal.Nutrition.Calories
	nutrition.Protein += meal.Nutrition.Protein
	nutrition.Carbs += meal.Nutrition.Carbs
	nutrition.Fat += meal.Nutrition.Fat
	nutrition.Sugar += meal.Nutrition.Sugar
	nutrition.Fiber += meal.Nutrition.Fiber
	nutrition.Sodium += meal.Nutrition.Sodium
	nutrition.Calcium += meal.Nutrition.Calcium
	nutrition.Iron += meal.Nutrition.Iron
	nutrition.Cholesterol += meal.Nutrition.Cholesterol
	nutrition.Potassium += meal.Nutrition.Potassium
	nutrition.VitaminA += meal.Nutrition.VitaminA
	nutrition.VitaminC += meal.Nutrition.VitaminC

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"user": userID, "date": date},
		bson.M{
			"$set":  bson.M{"nutrition": nutrition},
			"$push": bson.M{"meals": meal},
		},
	)
	if err != nil {
		fmt.Println("update failed!")
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to add meal into day collection:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
