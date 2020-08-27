package main

import (
	"context"
	"encoding/json"

	"../../lib"

	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	dateLayout = "2020-01-12"
)

type NutritionSummary struct {
	Calories int
}

type Food struct {
	Name      string
	Group     string
	Serving   string
	Nutrition NutritionSummary
}

type InputMeal struct {
	Name  string
	Foods []Food
}

type InputDayRecord struct {
	Date  string
	User  string
	Meals []InputMeal
}

type SavedMeal struct {
	Name      string
	Foods     []Food
	Nutrition NutritionSummary
}

type SavedDayRecord struct {
	Date      string
	User      string
	Meals     []SavedMeal
	Nutrition NutritionSummary
}

type getDayReq struct {
	User string
}

func main() {
	log.Println("Day Service start")

	http.Handle("/days", lib.CorsMiddleware(http.HandlerFunc(Days)))
	http.Handle("/day", lib.CorsMiddleware(http.HandlerFunc(Day)))

	log.Fatal(http.ListenAndServe(":8083", nil))
}

func Days(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Days")
	userID := r.URL.Query().Get("user")

	switch r.Method {
	case http.MethodGet:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cur, err := collection.Find(ctx, bson.M{"user": userID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("could not find day results:\n" + err.Error()))
			return
		}
		defer cur.Close(ctx)

		dayRecords := make([]SavedDayRecord, 0)

		for cur.Next(ctx) {
			var dayRecord SavedDayRecord
			err := cur.Decode(&dayRecord)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			dayRecords = append(dayRecords, dayRecord)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dayRecords)
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	}
}

func Day(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Days")

	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var req InputDayRecord
		err := decoder.Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("could not decode post day request:\n" + err.Error()))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dayCalories := 0

		reqMeals := bson.A{}
		for _, meal := range req.Meals {
			mealCalories := 0
			for _, food := range meal.Foods {
				mealCalories += food.Nutrition.Calories
			}
			savedMeal := SavedMeal{
				Name:      meal.Name,
				Foods:     meal.Foods,
				Nutrition: NutritionSummary{mealCalories},
			}
			reqMeals = append(reqMeals, savedMeal)
			dayCalories += mealCalories
		}

		dayNutrition := NutritionSummary{dayCalories}

		res, err := collection.InsertOne(ctx, bson.D{
			{Key: "user", Value: req.User},
			{Key: "date", Value: req.Date},
			{Key: "meals", Value: reqMeals},
			{Key: "nutrition", Value: dayNutrition},
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to insert into day collection:\n" + err.Error()))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
	}
}
