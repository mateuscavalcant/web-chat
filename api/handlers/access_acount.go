package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"web-chat/api/utils"
	"web-chat/pkg/database"
	"web-chat/pkg/models"
	"web-chat/pkg/models/err"

	"github.com/patrickmn/go-cache"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
    loginAttempts = cache.New(5*time.Minute, 10*time.Minute)
)

// AccessUserAccount handles user authentication.
func AccessUserAccount(c *gin.Context) {
	// Define a struct to hold user information.
	var user models.User

	// Extract email and password from the POST form data.
	email := strings.TrimSpace(c.PostForm("email"))
	password := strings.TrimSpace(c.PostForm("password"))

	// Create a response object to handle errors.
	resp := err.ErrorResponse{
		Error: make(map[string]string),
	}

	// Get the database connection from the pool.
	db := database.GetDB()

	// Query the database to retrieve user information.
	row := db.QueryRow("SELECT id, email, password FROM user WHERE email=?", email)
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		// Log the error and set appropriate response for invalid credentials.
		log.Println("Error executing SQL statement:", err)
		resp.Error["credentials"] = "Invalid credentials"
	
		if attempts, found := loginAttempts.Get(email); found {
			loginAttempts.Set(email, attempts.(int)+1, cache.DefaultExpiration)
		} else {
			loginAttempts.Set(email, 1, cache.DefaultExpiration)
		}
	
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	
	// Compare the provided password with the hashed password retrieved from the database.
	encErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if encErr != nil {
		// Set appropriate response for invalid password.
		resp.Error["password"] = "Invalid password"
	
		// Incrementa o contador de tentativas de login para este usuário.
		if attempts, found := loginAttempts.Get(email); found {
			loginAttempts.Set(email, attempts.(int)+1, cache.DefaultExpiration)
		} else {
			loginAttempts.Set(email, 1, cache.DefaultExpiration)
		}
	
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	
	// If authentication is successful, remove the login attempts counter for this user.
	loginAttempts.Delete(email)
	
	// If authentication is successful, store user information in session and return success message.
	session := utils.GetSession(c)
	session.Values["id"] = strconv.Itoa(user.ID)
	session.Values["email"] = user.Email
	session.Save(c.Request, c.Writer)
	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})

}