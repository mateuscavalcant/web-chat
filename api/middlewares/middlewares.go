package middlewares

import (
	"log"
	"net/http"
	"time"
	"web-chat/api/utils"

	"github.com/patrickmn/go-cache"
)

var loginAttempts *cache.Cache

func init() {
    loginAttempts = cache.New(5*time.Minute, 10*time.Minute)
}

func LimitLoginAttempts(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        email := r.PostFormValue("email")

        if attempts, found := loginAttempts.Get(email); found && attempts.(int) >= 3 {
            http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}


// AuthMiddleware is a middleware function to authenticate user sessions.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve session information from the request.
		session := utils.GetSession(r)
		userID := session.Values["id"]

		// Check if the user ID is not present in the session.
		if userID == nil {
			// If not authorized.
			log.Println("Unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// If user is authenticated, proceed to the next handler.
		next.ServeHTTP(w, r)
	})
}
