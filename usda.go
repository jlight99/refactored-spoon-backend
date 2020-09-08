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

type FoodSearchRequest struct {
	Food     string `json:"food,omitempty"`
	PageSize string `json:"pageSize,omitempty"`
}

type FoodSearchCriteria struct {
	GeneralSearchInput string `json:"generalSearchInput,omitempty"`
	PageNumber         int    `json:"pageNumber,omitempty"`
	RequireAllWords    bool   `json:"requireAllWords,omitempty"`
}

type UsdaFood struct {
	FdcId       int    `json:"fdcId,omitempty"`
	Description string `json:"description,omitempty"`
	BrandOwner  string `json:"brandOwner,omitempty"`
	Ingredients string `json:"ingredients,omitempty"`
}

type FoodSearchResult struct {
	FoodSearchCriteria FoodSearchCriteria `json:"foodSearchCriteria,omitempty"`
	CurrentPage        int                `json:"currentPage,omitempty"`
	TotalPages         int                `json:"totalPages,omitempty"`
	Foods              []UsdaFood         `json:"foods,omitempty"`
}

type FoodDetailRequest struct {
	Food int `json:"food,omitempty"`
}

type FoodsDetailRequest struct {
	Foods []int `json:"foods,omitempty"`
}

type LabelNutrient struct {
	Value float64 `json:"value,omitempty"`
}

// type LabelNutrients struct {
// 	Fat           LabelNutrient `json:"fat,omitempty"`
// 	SaturatedFat  LabelNutrient `json:"saturatedFat,omitempty"`
// 	TransFat      LabelNutrient `json:"transFat,omitempty"`
// 	Cholesterol   LabelNutrient `json:"cholesterol,omitempty"`
// 	Sodium        LabelNutrient `json:"sodium,omitempty"`
// 	Carbohydrates LabelNutrient `json:"carbohydrates,omitempty"`
// 	Fiber         LabelNutrient `json:"fiber,omitempty"`
// 	Sugars        LabelNutrient `json:"sugars,omitempty"`
// 	Protein       LabelNutrient `json:"protein,omitempty"`
// 	Calcium       LabelNutrient `json:"calcium,omitempty"`
// 	Iron          LabelNutrient `json:"iron,omitempty"`
// 	Calories      LabelNutrient `json:"calories,omitempty"`
// }

type Nutrient struct {
	Id       int    `json:"id,omitempty"`
	Number   string `json:"number,omitempty"`
	Name     string `json:"name,omitempty"`
	Rank     int    `json:"rank,omitempty"`
	UnitName string `json:"unitName,omitempty"`
}

type FoodNutrientDerivation struct {
	Id                 int                `json:"id,omitempty"`
	Code               string             `json:"code,omitempty"`
	Description        string             `json:"description,omitempty"`
	FoodNutrientSource FoodNutrientSource `json:"foodNutrientSource,omitempty"`
}

type FoodNutrientSource struct {
	Id          int    `json:"id,omitempty"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

type FoodNutrient struct {
	Type     string   `json:"type,omitempty"`
	Id       int      `json:"id,omitempty"`
	Nutrient Nutrient `json:"nutrient,omitempty"`
	Amount   float64  `json:"amount,omitempty"`
}

type FoodDetailResult struct {
	FdcId           int            `json:"fdcId,omitempty"`
	FoodClass       string         `json:"foodClass,omitempty"`
	Description     string         `json:"description,omitempty"`
	Ingredients     string         `json:"ingredients,omitempty"`
	ServingSize     float64        `json:"servingSize,omitempty"`
	ServingSizeUnit string         `json:"servingSizeUnit,omitempty"`
	FoodNutrients   []FoodNutrient `json:"foodNutrients,omitempty"`
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
