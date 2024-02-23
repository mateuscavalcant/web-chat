package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

// store holds the session store.
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

// GetSession retrieves the session from the request.
func GetSession(r *http.Request) *sessions.Session {
	// Retrieve the session from the store.
	session, err := store.Get(r, "session")
	// Handle any errors.
	Err(err)
	// Return the session.
	return session
}

// AllSessions retrieves all session data.
func AllSessions(r *http.Request) (interface{}, interface{}) {
	// Retrieve the session.
	session := GetSession(r)
	// Extract id and email from the session.
	id := session.Values["id"]
	email := session.Values["email"]
	// Return id and email.
	return id, email
}

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
