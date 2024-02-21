package utils

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

// store holds the session store.
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

// GetSession retrieves the session from the request.
func GetSession(c *gin.Context) *sessions.Session {
	// Retrieve the session from the store.
	session, err := store.Get(c.Request, "session")
	// Handle any errors.
	Err(err)
	// Return the session.
	return session
}

// AllSessions retrieves all session data.
func AllSessions(c *gin.Context) (interface{}, interface{}) {
	// Retrieve the session.
	session := GetSession(c)
	// Extract id and email from the session.
	id := session.Values["id"]
	email := session.Values["email"]
	// Return id and email.
	return id, email
}

// Err logs fatal errors.
func Err(err interface{}) {
	if err != nil {
		log.Fatal(err)
	}
}

// Ses returns session information.
func Ses(c *gin.Context) interface{} {
	// Retrieve session information.
	id, username := AllSessions(c)
	// Return session data as a map.
	return map[string]interface{}{
		"id":       id,
		"username": username,
	}
}