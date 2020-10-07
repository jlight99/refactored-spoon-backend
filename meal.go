package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/internal/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MealsHandler handles /meals GET and POST requests
func MealsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	collection := lib.GetCollection("Days")
	userID := r.URL.Query().Get("userId")

	switch r.Method {
	case http.MethodGet:
		getMeals(w, r, collection, userID, date)
	case http.MethodPost:
		postMeal(w, r, collection, userID, date)
	}
}

// MealHandler handles /meals/{mealId} PUT and DELETE requests
func MealHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]
	mealID := vars["mealId"]
	mealObjectID, err := primitive.ObjectIDFromHex(mealID)
	if err != nil {
		fmt.Println("error: invalid meal ID provided: " + mealID)
		return
	}

	collection := lib.GetCollection("Days")
	userID := r.URL.Query().Get("userId")

	switch r.Method {
	case http.MethodDelete:
		deleteMeal(w, r, collection, userID, date, mealObjectID)
	case http.MethodPut:
		updateMeal(w, r, collection, userID, date, mealObjectID)
	}
}

func getMeals(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{"userId": userID, "date": date})
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

	// generate ids for meal and food nested objects
	meal.ID = primitive.NewObjectID()
	for i, _ := range meal.Foods {
		meal.Foods[i].ID = primitive.NewObjectID()
	}

	// get day record or create one if it doesn't exist
	dayRecord := GetDayByDate(ctx, collection, userID, date)
	if dayRecord.ID == primitive.NilObjectID {
		dayRecord = &DayRecord{
			Date:      date,
			UserID:    userID,
			Meals:     []Meal{},
			Nutrition: NutritionSummary{},
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

	// add meal nutrition to day record nutrition
	nutrition := updateNutrition(dayRecord.Nutrition, meal.Nutrition, 1.0)

	// update day record with new meal
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"userId": userID, "date": date},
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

func updateMeal(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string, mealID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	decoder := json.NewDecoder(r.Body)
	var meal Meal
	err := decoder.Decode(&meal)
	if err != nil {
		fmt.Println("could not decode post meal request:\n" + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode post meal request:\n" + err.Error()))
		return
	}

	// generate id for new foods
	for i, _ := range meal.Foods {
		if meal.Foods[i].ID == primitive.NilObjectID {
			meal.Foods[i].ID = primitive.NewObjectID()
		}
	}

	dayRecord := GetDayByDate(ctx, collection, userID, date)
	meals := dayRecord.Meals

	// find original value for meal to be updated
	var originalMealIdx int
	for i, _ := range meals {
		if meals[i].ID == mealID {
			originalMealIdx = i
			break
		}
	}

	// replace original meal's nutrition with the updated meal's nutrition in the total day nutrition
	nutrition := updateNutrition(dayRecord.Nutrition, meals[originalMealIdx].Nutrition, -1.0) // subtract original nutrition
	nutrition = updateNutrition(nutrition, meal.Nutrition, 1.0)                               // add new nutrition

	// replace original meal with new meal
	meals[originalMealIdx] = meal

	// commit changes to database
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"userId": userID, "date": date},
		bson.M{
			"$set": bson.M{"nutrition": nutrition, "meals": meals},
		},
	)
	if err != nil {
		fmt.Println("unable to add meal into day collection:\n" + err.Error())
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
		bson.M{"userId": userID, "date": date, "meals": bson.M{"$elemMatch": bson.M{"_id": mealID}}},
		bson.M{
			"$set":  bson.M{"nutrition": updateNutrition(dayRecord.Nutrition, mealToDelete.Nutrition, -1)},
			"$pull": bson.M{"meals": mealToDelete},
		},
	)
}

func updateNutrition(dayNutrition NutritionSummary, mealNutrition NutritionSummary, sign float64) NutritionSummary {
	nutrition := dayNutrition
	updateNutrient(&nutrition.Calories, mealNutrition.Calories, sign)
	updateNutrient(&nutrition.Protein, mealNutrition.Protein, sign)
	updateNutrient(&nutrition.Carbs, mealNutrition.Carbs, sign)
	updateNutrient(&nutrition.Fat, mealNutrition.Fat, sign)
	updateNutrient(&nutrition.Sugar, mealNutrition.Sugar, sign)
	updateNutrient(&nutrition.Fiber, mealNutrition.Fiber, sign)
	updateNutrient(&nutrition.Sodium, mealNutrition.Sodium, sign)
	updateNutrient(&nutrition.Calcium, mealNutrition.Calcium, sign)
	updateNutrient(&nutrition.Iron, mealNutrition.Iron, sign)
	updateNutrient(&nutrition.Cholesterol, mealNutrition.Cholesterol, sign)
	updateNutrient(&nutrition.Potassium, mealNutrition.Potassium, sign)
	updateNutrient(&nutrition.VitaminA, mealNutrition.VitaminA, sign)
	updateNutrient(&nutrition.VitaminC, mealNutrition.VitaminC, sign)

	return nutrition
}

func updateNutrient(dayNutrient *Nutrient, mealNutrient Nutrient, sign float64) {
	dayNutrient.NutrientName = mealNutrient.NutrientName
	dayNutrient.UnitName = mealNutrient.UnitName
	dayNutrient.Value += sign * mealNutrient.Value
}
