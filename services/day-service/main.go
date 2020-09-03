package main

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/refactored-spoon-backend/lib"

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
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	}
}

func sortMeals(meals []Meal) {
	sort.Slice(meals, func(i, j int) bool {
		if meals[i].Name == "dinner" || meals[j].Name == "breakfast" {
			return false
		}
		if meals[i].Name == "breakfast" || meals[j].Name == "dinner" {
			return true
		}
		if meals[i].Name == "lunch" && meals[j].Name == "dinner" {
			return true
		}
		return false
	})
}

func Day(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Days")

	switch r.Method {
	case http.MethodPost:
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
	case http.MethodPut:
		decoder := json.NewDecoder(r.Body)
		var dayRecord DayRecord
		err := decoder.Decode(&dayRecord)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("could not decode put day request:\n" + err.Error()))
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

		_, err = collection.ReplaceOne(ctx, bson.M{"user": dayRecord.User, "date": dayRecord.Date}, dayRecord)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to insert into day collection:\n" + err.Error()))
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
