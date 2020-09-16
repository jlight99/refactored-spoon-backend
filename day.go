package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/internal/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NutritionSummary struct {
	Calories    Nutrient `json:"calories,omitempty" bson:"calories,omitempty"`
	Protein     Nutrient `json:"protein,omitempty" bson:"protein,omitempty"`
	Carbs       Nutrient `json:"carbs,omitempty" bson:"carbs,omitempty"`
	Fat         Nutrient `json:"fat,omitempty" bson:"fat,omitempty"`
	Sugar       Nutrient `json:"sugar,omitempty" bson:"sugar,omitempty"`
	Fiber       Nutrient `json:"fiber,omitempty" bson:"fiber,omitempty"`
	Sodium      Nutrient `json:"sodium,omitempty" bson:"sodium,omitempty"`
	Calcium     Nutrient `json:"calcium,omitempty" bson:"calcium,omitempty"`
	Iron        Nutrient `json:"iron,omitempty" bson:"iron,omitempty"`
	Cholesterol Nutrient `json:"cholesterol,omitempty" bson:"cholesterol,omitempty"`
	Potassium   Nutrient `json:"potassium,omitempty" bson:"potassium,omitempty"`
	VitaminA    Nutrient `json:"vitaminA,omitempty" bson:"vitaminA,omitempty"`
	VitaminC    Nutrient `json:"vitaminC,omitempty" bson:"vitaminC,omitempty"`
}

type Nutrient struct {
	NutrientName string  `json:"nutrientName,omitempty" bson:"nutrientName,omitempty"`
	UnitName     string  `json:"unitName,omitempty" bson:"unitName,omitempty"`
	Value        float64 `json:"value,omitempty" bson:"value,omitempty"`
}

type Food struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name          string             `json:"name,omitempty" bson:"name,omitempty"`
	Group         string             `json:"group,omitempty" bson:"group,omitempty"`
	Serving       int                `json:"serving,omitempty" bson:"serving,omitempty"`
	Nutrition     NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`         // based on serving size
	USDANutrition NutritionSummary   `json:"usdaNutrition,omitempty" bson:"usdaNutrition,omitempty"` // source of truth, based on nutrients / 100 g
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
	UserID    string             `json:"userId,omitempty" bson:"userId,omitempty"`
	Meals     []Meal             `json:"meals,omitempty" bson:"meals,omitempty"`
	Nutrition NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
}

func DaysHandler(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Days")
	userID := r.URL.Query().Get("userId")

	switch r.Method {
	case http.MethodGet:
		getDays(w, r, collection, userID)
	case http.MethodPost:
		postDays(w, r, collection)
	}
}

func DayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	collection := lib.GetCollection("Days")
	userID := r.URL.Query().Get("userId")

	switch r.Method {
	case http.MethodGet:
		getDay(w, r, collection, userID, date)
	case http.MethodDelete:
		deleteDay(w, r, collection, userID, date)
	}
}

func getDays(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not find day results:\n" + err.Error()))
		return
	}
	defer cur.Close(ctx)

	dayRecords := make([]DayRecord, 0)

	for cur.Next(ctx) {
		var dayRecord DayRecord
		err := cur.Decode(&dayRecord)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		dayRecords = append(dayRecords, dayRecord)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dayRecords)
}

func postDays(w http.ResponseWriter, r *http.Request, collection *mongo.Collection) {
	decoder := json.NewDecoder(r.Body)
	var dayRecord DayRecord
	err := decoder.Decode(&dayRecord)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode post day request:\n" + err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sortMeals(dayRecord.Meals)

	for i, _ := range dayRecord.Meals {
		if dayRecord.Meals[i].ID == primitive.NilObjectID {
			dayRecord.Meals[i].ID = primitive.NewObjectID()
		}
		for j, _ := range dayRecord.Meals[i].Foods {
			if dayRecord.Meals[i].Foods[j].ID == primitive.NilObjectID {
				dayRecord.Meals[i].Foods[j].ID = primitive.NewObjectID()
			}
		}
	}

	res, err := collection.InsertOne(ctx, dayRecord)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to insert into day collection:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
}

func GetDayByDate(ctx context.Context, collection *mongo.Collection, userID string, date string) *DayRecord {
	var dayRecord DayRecord
	err := collection.FindOne(ctx, bson.M{"userId": userID, "date": date}).Decode(&dayRecord)
	if err != nil {
		log.Println("error in finding day with userId: " + userID + " date: " + date)
		log.Println(err)
		return &DayRecord{}
	}
	return &dayRecord
}

func getDay(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dayRecord := GetDayByDate(ctx, collection, userID, date)
	sortMeals(dayRecord.Meals)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dayRecord)
}

func deleteDay(w http.ResponseWriter, r *http.Request, collection *mongo.Collection, userID string, date string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection.FindOneAndDelete(ctx, bson.M{"userId": userID, "date": date})
}

func sortMeals(meals []Meal) {
	sort.Slice(meals, func(i, j int) bool {
		if meals[i].Name == "Dinner" || meals[j].Name == "Breakfast" {
			return false
		}
		if meals[i].Name == "Breakfast" || meals[j].Name == "Dinner" {
			return true
		}
		if meals[i].Name == "Lunch" && meals[j].Name == "Dinner" {
			return true
		}
		return false
	})
}
