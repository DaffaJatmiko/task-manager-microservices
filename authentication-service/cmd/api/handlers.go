package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DaffaJatmiko/authentication-service/data"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("jwt-key")

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func (app *Config) GenerateJWT(email string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return tokenString, nil

}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	log.Println("Authenticate users")
	
	// validate the user against the database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials by email") , http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		log.Println("Checking password of", requestPayload.Password)
		app.errorJSON(w, errors.New("invalid credentials by password"), http.StatusBadRequest)
		return
	}

	// log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// create a jwt token
	tokenString, err := app.GenerateJWT(user.Email)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    map[string]string{"token": tokenString},
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, err := json.MarshalIndent(entry, "", "\t")
	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) Register(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Password  string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	log.Println("Register user")

	// create a new user
	user := data.User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password,
		Active:    1, // setting user as active by default
		CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
	}

	// save the user to the database
	_, err = app.Models.User.Insert(user)
	if err != nil {
		app.errorJSON(w, errors.New("unable to create user"), http.StatusBadRequest)
		return
	}

	log.Printf("Email: %s, FirstName: %s, LastName: %s, Active: %d\n", user.Email, user.FirstName, user.LastName, user.Active)

	// log registration
	err = app.logRequest("registration", fmt.Sprintf("%s registered", user.Email))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Registered user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusCreated, payload)
}
