package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/refactored-spoon-backend/internal/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type userRequest struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  = "potatoe"
)

func HandleMain(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandleMain")
	var htmlIndex = `<html>
<body>
	<a href="/google-login">Google Log In</a>
</body>
</html>`
	fmt.Fprintf(w, htmlIndex)
}

func GoogleLogin() {
	fmt.Println("GoogleLogin")

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8081/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandleGoogleLogin")

	GoogleLogin()
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	fmt.Printf("redirect url: %s\n", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func getUserInfo(state string, code string) ([]byte, error) {
	fmt.Println("getUserInfo")

	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state: %s\n", state)
	}
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	fmt.Println("got token!")
	fmt.Println(token)
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	return contents, nil
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandleGoogleCallback")
	collection := lib.GetCollection("Users")

	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	contentMap := make(map[string]string)
	err = json.Unmarshal(content, &contentMap)
	email := contentMap["email"]

	// fmt.Fprintf(w, "Content: %s\n", content)

	var user userRequest
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		fmt.Printf("couldn't find user with this google email: %s\n", email)
		res, err := collection.InsertOne(ctx, bson.M{"email": email, "password": "google"})
		if err != nil {
			fmt.Println("couldn't insert new user from google signin")
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Printf("returning this: %s", []byte(res.InsertedID.(primitive.ObjectID).Hex()))
		w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
	}

	w.WriteHeader(http.StatusOK)
	fmt.Printf("returning this: %s", user.ID.Hex())
	w.Write([]byte(user.ID.Hex()))
}

func Signup(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Users")

	decoder := json.NewDecoder(r.Body)
	var userReq userRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode user signup request:\n" + err.Error()))
		return
	}

	// verify on client side that both email and password are provided

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, bson.M{"email": userReq.Email, "password": userReq.Password})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to insert into user collection:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res.InsertedID.(primitive.ObjectID).Hex()))
}

func Login(w http.ResponseWriter, r *http.Request) {
	collection := lib.GetCollection("Users")

	decoder := json.NewDecoder(r.Body)
	var userReq userRequest
	err := decoder.Decode(&userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not decode user login request:\n" + err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := collection.FindOne(ctx, bson.M{"email": userReq.Email, "password": userReq.Password})

	var findRes userRequest
	err = res.Decode(&findRes)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("could not find user with this login:\n" + err.Error()))
		return
	}

	w.WriteHeader(http.StatusFound)
	w.Write([]byte(findRes.ID.Hex()))
}
