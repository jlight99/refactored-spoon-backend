package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/refactored-spoon-backend/internal/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NutritionSummary contains information about all the nutrients
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

// Nutrient contains information about a single nutrient
type Nutrient struct {
	NutrientName string  `json:"nutrientName,omitempty" bson:"nutrientName,omitempty"`
	UnitName     string  `json:"unitName,omitempty" bson:"unitName,omitempty"`
	Value        float64 `json:"value,omitempty" bson:"value,omitempty"`
}

// Food contains information such as name, group, serving size, and nutrition
type Food struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name          string             `json:"name,omitempty" bson:"name,omitempty"`
	Group         string             `json:"group,omitempty" bson:"group,omitempty"`
	Serving       int                `json:"serving,omitempty" bson:"serving,omitempty"`
	Nutrition     NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`         // based on serving size
	USDANutrition NutritionSummary   `json:"usdaNutrition,omitempty" bson:"usdaNutrition,omitempty"` // source of truth, based on nutrients / 100 g
}

// Meal contains the type of meal, a list of foods, and the nutrition summary of the meal
type Meal struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty"`
	Foods     []Food             `json:"foods,omitempty" bson:"foods,omitempty"`
	Nutrition NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
}

// DayRecord is the representation of all the foods a user ate during a day as well as a nutrition summary
type DayRecord struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Date      string             `json:"date,omitempty" bson:"date,omitempty"`
	UserID    string             `json:"userId,omitempty" bson:"userId,omitempty"`
	Meals     []Meal             `json:"meals,omitempty" bson:"meals,omitempty"`
	Nutrition NutritionSummary   `json:"nutrition,omitempty" bson:"nutrition,omitempty"`
}

// used to sort meals
var mealValues = map[string]int{
	"breakfast": 1,
	"lunch":     2,
	"dinner":    3,
}

// DayHandler handles /days/{dayId} GET and DELETE requests
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

// used in day.go:getDay, meal.go:postMeal, updateMeal, deleteMeal
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

func sortMeals(meals []Meal) {
	sort.Slice(meals, func(i, j int) bool {
		mealName1 := strings.ToLower(meals[i].Name)
		mealName2 := strings.ToLower(meals[j].Name)
		return mealValues[mealName1] < mealValues[mealName2]
	})
}
