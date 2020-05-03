package main

import (
	"bytes"
	"encoding/json"
	"github.com/refactored-spoon-backend/lib"
	"log"
	"net/http"
)

const(
	apiKey = "EH6jWPD9LdlAPYrnQzR4luccqsnhUBwSd99kwocV"
	usdaFoodDataCentralEndpoint = "https://api.nal.usda.gov/fdc/v1/"
)

type foodSearchRequest struct {
	Food string
}

type foodSearchCriteria struct {
	GeneralSearchInput string
	PageNumber int
	RequireAllWords bool
}

type usdaFood struct {
	FdcId int
	Description string
	BrandOwner string
	Ingredients string
}

type foodSearchResult struct {
	FoodSearchCriteria foodSearchCriteria
	CurrentPage int
	TotalPages int
	Foods []usdaFood
}

type foodDetailRequest struct {
	Food string
}

type labelNutrient struct {
	Value float64
}

type labelNutrients struct {
	Fat labelNutrient
	SaturatedFat labelNutrient
	TransFat labelNutrient
	Cholesterol labelNutrient
	Sodium labelNutrient
	Carbohydrates labelNutrient
	Fiber labelNutrient
	Sugars labelNutrient
	Protein labelNutrient
	Calcium labelNutrient
	Iron labelNutrient
	Calories labelNutrient
}

type foodDetailResult struct {
	FoodClass string
	Description string
	Ingredients string
	ServingSize float64
	ServingSizeUnit string
	LabelNutrients labelNutrients
}

func main() {
	log.Println("USDA Service start")

	http.Handle("/food/search", lib.CorsMiddleware(http.HandlerFunc(SearchFood)))
	http.Handle("/food/detail", lib.CorsMiddleware(http.HandlerFunc(FoodDetail)))

	log.Fatal(http.ListenAndServe(":8082", nil))
}

func SearchFood(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	decoder := json.NewDecoder(r.Body)
	var queryStr foodSearchRequest
	err := decoder.Decode(&queryStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food search request"))
		return
	}

	searchReqBody := []byte(`{"generalSearchInput":"` + queryStr.Food + `"}`)

	req, err := http.NewRequest(http.MethodPost, usdaFoodDataCentralEndpoint + "search?api_key=" + apiKey, bytes.NewBuffer(searchReqBody))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to create POST request to search USDA food data central db"))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to send POST request to search USDA food data central db"))
		return
	}
	defer res.Body.Close()

	decoder = json.NewDecoder(res.Body)
	var searchResults foodSearchResult
	err = decoder.Decode(&searchResults)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food search results: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResults.Foods)
}

func FoodDetail(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	decoder := json.NewDecoder(r.Body)
	var queryStr foodDetailRequest
	err := decoder.Decode(&queryStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food detail request"))
		return
	}

	req, err := http.NewRequest(http.MethodGet, usdaFoodDataCentralEndpoint + queryStr.Food + "?api_key=" + apiKey, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to create GET request to detail USDA food data central db"))
		return
	}

	res, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to send GET request to detail USDA food data central db"))
		return
	}
	defer res.Body.Close()

	decoder = json.NewDecoder(res.Body)
	var searchResults foodDetailResult
	err = decoder.Decode(&searchResults)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food detail results: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResults)
}
