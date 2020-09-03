package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/refactored-spoon-backend/lib"

	"log"
	"net/http"
)

const (
	apiKey                      = "EH6jWPD9LdlAPYrnQzR4luccqsnhUBwSd99kwocV"
	usdaFoodDataCentralEndpoint = "https://api.nal.usda.gov/fdc/v1/"
)

type FoodSearchRequest struct {
	Food     string
	PageSize string
}

type FoodSearchCriteria struct {
	GeneralSearchInput string
	PageNumber         int
	RequireAllWords    bool
}

type UsdaFood struct {
	FdcId       int
	Description string
	BrandOwner  string
	Ingredients string
}

type FoodSearchResult struct {
	FoodSearchCriteria FoodSearchCriteria
	CurrentPage        int
	TotalPages         int
	Foods              []UsdaFood
}

type FoodDetailRequest struct {
	Food int
}

type FoodsDetailRequest struct {
	Foods []int
}

type LabelNutrient struct {
	Value float64
}

type LabelNutrients struct {
	Fat           LabelNutrient
	SaturatedFat  LabelNutrient
	TransFat      LabelNutrient
	Cholesterol   LabelNutrient
	Sodium        LabelNutrient
	Carbohydrates LabelNutrient
	Fiber         LabelNutrient
	Sugars        LabelNutrient
	Protein       LabelNutrient
	Calcium       LabelNutrient
	Iron          LabelNutrient
	Calories      LabelNutrient
}

type Nutrient struct {
	Id       int
	Number   string
	Name     string
	Rank     int
	UnitName string
}

type FoodNutrientDerivation struct {
	Id                 int
	Code               string
	Description        string
	FoodNutrientSource FoodNutrientSource
}

type FoodNutrientSource struct {
	Id          int
	Code        string
	Description string
}

type FoodNutrient struct {
	Type     string
	Id       int
	Nutrient Nutrient
	Amount   float64
}

type FoodDetailResult struct {
	FdcId           int
	FoodClass       string
	Description     string
	Ingredients     string
	ServingSize     float64
	ServingSizeUnit string
	// LabelNutrients  LabelNutrients
	FoodNutrients []FoodNutrient
}

func main() {
	log.Println("USDA Service start")

	http.Handle("/food/search", lib.CorsMiddleware(http.HandlerFunc(SearchFood)))
	http.Handle("/food/detail", lib.CorsMiddleware(http.HandlerFunc(FoodDetail)))
	http.Handle("/foods/detail", lib.CorsMiddleware(http.HandlerFunc(FoodsDetail)))

	log.Fatal(http.ListenAndServe(":8082", nil))
}

func SearchFood(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	decoder := json.NewDecoder(r.Body)
	var queryStr FoodSearchRequest
	err := decoder.Decode(&queryStr)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food search request:\n" + err.Error()))
		return
	}

	searchReqBody := []byte(`{"generalSearchInput":"` + queryStr.Food + `", "pageSize":` + queryStr.PageSize + `}`)

	req, err := http.NewRequest(http.MethodPost, usdaFoodDataCentralEndpoint+"search?api_key="+apiKey, bytes.NewBuffer(searchReqBody))
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to create POST request to search USDA food data central db:\n" + err.Error()))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to send POST request to search USDA food data central db:\n" + err.Error()))
		return
	}
	defer res.Body.Close()

	decoder = json.NewDecoder(res.Body)
	var searchResults FoodSearchResult
	err = decoder.Decode(&searchResults)
	if err != nil {
		log.Println(err.Error())
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
	var queryStr FoodDetailRequest
	err := decoder.Decode(&queryStr)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food detail request:\n" + err.Error()))
		return
	}

	req, err := http.NewRequest(http.MethodGet, usdaFoodDataCentralEndpoint+strconv.Itoa(queryStr.Food)+"?api_key="+apiKey, nil)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to create GET request to detail USDA food data central db:\n" + err.Error()))
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to send GET request to detail USDA food data central db:\n" + err.Error()))
		return
	}
	defer res.Body.Close()

	decoder = json.NewDecoder(res.Body)
	var searchResults FoodDetailResult
	err = decoder.Decode(&searchResults)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food detail results: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResults)
}

func FoodsDetail(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	decoder := json.NewDecoder(r.Body)
	var queryStr FoodsDetailRequest
	err := decoder.Decode(&queryStr)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode foods detail request:\n" + err.Error()))
		return
	}

	fdcIds := ""

	for _, fdcId := range queryStr.Foods {
		fdcIds += strconv.Itoa(fdcId) + ","
	}
	fdcIds = strings.TrimSuffix(fdcIds, ",")

	req, err := http.NewRequest(http.MethodGet, usdaFoodDataCentralEndpoint+"foods?api_key="+apiKey+"&fdcIds="+fdcIds, nil)
	// req, err := http.NewRequest(http.MethodGet, usdaFoodDataCentralEndpoint+strconv.Itoa(queryStr.Food)+"?api_key="+apiKey, nil)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to create GET request to detail USDA food data central db:\n" + err.Error()))
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to send GET request to detail USDA food data central db:\n" + err.Error()))
		return
	}
	defer res.Body.Close()

	decoder = json.NewDecoder(res.Body)
	var searchResults []FoodDetailResult
	err = decoder.Decode(&searchResults)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode foods detail results: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResults)
}
