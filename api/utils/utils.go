package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Err logs fatal errors.
func Err(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Ses returns session information.
func Ses(r *http.Request) interface{} {
	// Retrieve session information.
	id, username := AllSessions(r)
	// Return session data as a map.
	return map[string]interface{}{
		"id":       id,
		"username": username,
	}
}

// respondWithError responds with the provided error message and status code.
func RespondWithError(w http.ResponseWriter, message interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(gin.H{"error": message})
}

// respondWithJSON responds with the provided data and status code.
func RespondWithJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
