package userapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/PhongVX/golang-rest-api/auth"
	"github.com/PhongVX/golang-rest-api/db"
	"github.com/PhongVX/golang-rest-api/entities"
)

// receive username - password from client -> authentication -> generate access token and POST request to client
func Authorize(response http.ResponseWriter, request *http.Request) {

	var tokenRequest entities.TokenRequest
	err := json.NewDecoder(request.Body).Decode(&tokenRequest)

	if err != nil {
		responseWithError(response, http.StatusForbidden, err.Error())
	}

	if tokenRequest.Grant_Type != "password" || !checkPermit(tokenRequest) {
		responseWithError(response, http.StatusForbidden, "Unauthorized")
	}
	os.Setenv("SECRET_KEY", tokenRequest.Client_Secret)
	token, err := auth.CreateToken(tokenRequest.Username, tokenRequest.Password)
	if err != nil {
		responseWithError(response, http.StatusForbidden, err.Error())
	}
	requestBody, err := json.Marshal(map[string]string{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   string(time.Now().Add(time.Hour * 1).Unix()),
	})
	if err != nil {
		responseWithError(response, http.StatusBadRequest, err.Error())
	}
	resp, err := http.Post("http://127.0.0.1:4000/api/resource", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		responseWithError(response, http.StatusBadRequest, err.Error())
	}
	responseWithJSON(response, http.StatusOK, token)
	defer resp.Body.Close()

}

// get access token from client -> return resource
func GetResource(response http.ResponseWriter, request *http.Request) {
	err, data := auth.TokenValid(request)
	if err != nil {
		responseWithError(response, http.StatusForbidden, err.Error())
	} else {
		resource := db.GetData(data.Username, data.Password)
		responseWithJSON(response, http.StatusOK, resource)
	}
}

// authenticate resource owner by access token
func checkPermit(request entities.TokenRequest) bool {
	user := db.GetData(request.Username, request.Password)
	if user.Username != "" {
		return true
	}
	return false

}

func responseWithError(response http.ResponseWriter, statusCode int, msg string) {
	responseWithJSON(response, statusCode, map[string]string{
		"error": msg,
	})
}

func responseWithJSON(response http.ResponseWriter, statusCode int, data interface{}) {
	result, _ := json.Marshal(data)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)
	response.Write(result)
}
