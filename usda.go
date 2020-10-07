package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	usdaFoodDataCentralEndpoint = "https://api.nal.usda.gov/fdc/v1/"
)

var (
	apiKey = os.Getenv("USDA_API_KEY")
)

// FoodSearchCriteria is the body for the USDA POST /search request
type FoodSearchCriteria struct {
	GeneralSearchInput string `json:"generalSearchInput,omitempty"`
	PageNumber         int    `json:"pageNumber,omitempty"`
	PageSize           int    `json:"pageSize,omitempty"`
	RequireAllWords    bool   `json:"requireAllWords,omitempty"`
}

// UsdaFood is the food result in the USDA POST /search response
type UsdaFood struct {
	FdcId         int                    `json:"fdcId,omitempty"`
	Description   string                 `json:"description,omitempty"`
	BrandOwner    string                 `json:"brandOwner,omitempty"`
	Ingredients   string                 `json:"ingredients,omitempty"`
	FoodNutrients []AbridgedFoodNutrient `json:"foodNutrients,omitempty"`
}

// AbridgedFoodNutrient is the nutrient result in the USDA POST /search response
// note that this contains much less information than FoodNutrient, which is what the USDA POST /foods response contains
type AbridgedFoodNutrient struct {
	NutrientId   int     `json:"nutrientId,omitempty"`
	NutrientName string  `json:"nutrientName,omitempty"`
	UnitName     string  `json:"unitName,omitempty"`
	Value        float64 `json:"value,omitempty"`
}

// FoodSearchResult is the body of the USDA POST /search response
type FoodSearchResult struct {
	FoodSearchCriteria FoodSearchCriteria `json:"foodSearchCriteria,omitempty"`
	CurrentPage        int                `json:"currentPage,omitempty"`
	TotalPages         int                `json:"totalPages,omitempty"`
	Foods              []UsdaFood         `json:"foods,omitempty"`
}

// Food Details (/foods) (currently unused)

// FoodDetailRequest is the body for the USDA POST /foods request
type FoodDetailRequest struct {
	FdcId int `json:"fdcId,omitempty"`
}

// FoodsDetailRequest is the body for the USDA bulk POST /foods request
type FoodsDetailRequest struct {
	FdcIds []int `json:"fdcIds,omitempty"`
}

// FoodDetailRequest is the body for the USDA POST /foods response
type FoodDetailResult struct {
	FdcId           int            `json:"fdcId,omitempty"`
	FoodClass       string         `json:"foodClass,omitempty"`
	Description     string         `json:"description,omitempty"`
	Ingredients     string         `json:"ingredients,omitempty"`
	ServingSize     float64        `json:"servingSize,omitempty"`
	ServingSizeUnit string         `json:"servingSizeUnit,omitempty"`
	FoodNutrients   []FoodNutrient `json:"foodNutrients,omitempty"`
}

// FoodNutrient is the nutrient result in the USDA POST /foods response
// note that this contains more information than the AbridgedFoodNutrient returned in the POST /search response
type FoodNutrient struct {
	Type     string       `json:"type,omitempty"`
	Id       int          `json:"id,omitempty"`
	Nutrient USDANutrient `json:"nutrient,omitempty"`
	Amount   float64      `json:"amount,omitempty"`
}

// USDANutrient is the USDA-specific nutrient result in the USDA POST /foods response
type USDANutrient struct {
	Id       int    `json:"id,omitempty"`
	Number   string `json:"number,omitempty"`
	Name     string `json:"name,omitempty"`
	Rank     int    `json:"rank,omitempty"`
	UnitName string `json:"unitName,omitempty"`
}

// End of Food Details

// SearchFood queries the USDA database by search keyword string and retrieves a list of matching foods with basic information
func SearchFood(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	decoder := json.NewDecoder(r.Body)
	var foodSearchCriteria FoodSearchCriteria
	err := decoder.Decode(&foodSearchCriteria)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to decode food search request:\n" + err.Error()))
		return
	}

	foodSearchCriteriaJSON, err := json.Marshal(foodSearchCriteria)
	if err != nil {
		log.Println(err.Error())
		return
	}

	req, err := http.NewRequest(http.MethodPost, usdaFoodDataCentralEndpoint+"search?api_key="+apiKey, bytes.NewBuffer(foodSearchCriteriaJSON))
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
	json.NewEncoder(w).Encode(searchResults)
}

// FoodDetail queries the USDA database and returns the details of a food given the FoodData Central ID of the food
// note that this function is currently not being used
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

	req, err := http.NewRequest(http.MethodGet, usdaFoodDataCentralEndpoint+strconv.Itoa(queryStr.FdcId)+"?api_key="+apiKey, nil)
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

// FoodsDetail queries the USDA database and returns the details of a list of foods given a list of FoodData Central IDs of the foods
// note that this function is currently not being used
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

	for _, fdcId := range queryStr.FdcIds {
		fdcIds += strconv.Itoa(fdcId) + ","
	}
	fdcIds = strings.TrimSuffix(fdcIds, ",")

	req, err := http.NewRequest(http.MethodGet, usdaFoodDataCentralEndpoint+"foods?api_key="+apiKey+"&fdcIds="+fdcIds, nil)
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
