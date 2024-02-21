package middlewares

import (
	"log"
	"net/http"
	"time"
	"web-chat/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)


func LimitLoginAttempts() gin.HandlerFunc {
	return func(c *gin.Context) {

		loginAttempts := cache.New(5*time.Minute, 10*time.Minute)
		username := c.PostForm("username") 

		if attempts, found := loginAttempts.Get(username); found && attempts.(int) >= 3 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Please try again later."})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthMiddleware is a middleware function to authenticate user sessions.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve session information from the request.
		session := utils.GetSession(c)
		userID := session.Values["id"]

		// Check if the user ID is not present in the session.
		if userID == nil {
			// If not authorized.
			log.Println("Unauthorized")
			c.JSON(http.StatusUnauthorized, gin.H{"message" : "Unauthorized"})
			c.Abort() // Abort further processing of the request.
			return
		}

		// If user is authenticated, proceed to the next handler.
		c.Next()
	}
}


