package main

import (
	"context"
	"encoding/json"
	"github.com/refactored-spoon-backend/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"time"
)

const (
	dateLayout = "2020-01-12"
)

type nutritionSummary struct {
	calories int
}

type food struct {
	name string
	group string
	serving string
	nutrition nutritionSummary
}

type meal struct {
	foods []food
	nutrition nutritionSummary
}

type meals struct {
	breakfast meal
	lunch meal
	dinner meal
}

type DayRecord struct {
	User primitive.ObjectID
	Date primitive.DateTime
	Meals meals
	Nutrition nutritionSummary
}

type TestDayRecord struct {
	User string
	Date string
}

type postDayReq struct {
	Date string
	User string
}

type getDayReq struct {
	User string
}

func main() {
	log.Println("Day Service start")

	collection := lib.GetCollection("Days")

	dayHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			decoder := json.NewDecoder(r.Body)
			var req getDayReq
			err := decoder.Decode(&req)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("could not decode get day request"))
				return
			}

			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			cur, err := collection.Find(ctx, bson.M{"user": req.User})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("could not find day results"))
				return
			}
			defer cur.Close(ctx)

			dayRecords := make([]TestDayRecord, 0)

			for cur.Next(ctx) {
				var dayRecord TestDayRecord
				err := cur.Decode(&dayRecord)
				if err != nil {
					return
				}
				dayRecords = append(dayRecords, dayRecord)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(dayRecords)
		case http.MethodPost:
			decoder := json.NewDecoder(r.Body)
			var req postDayReq
			err := decoder.Decode(&req)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("could not decode post day request"))
				return
			}

			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			res, err := collection.InsertOne(ctx, bson.M{"user": req.User, "date": req.Date})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("unable to insert into day collection"))
				return
			}

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
		case http.MethodPut:
		case http.MethodDelete:
		}
	}

	http.HandleFunc("/days", dayHandler)

	log.Fatal(http.ListenAndServe(":8083", nil))
}
