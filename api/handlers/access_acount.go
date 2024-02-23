package handlers

import (
	"encoding/json"
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
	"golang.org/x/crypto/bcrypt"
)

var (
	loginAttempts = cache.New(5*time.Minute, 10*time.Minute)
)

// AccessUserAccount handles user authentication.
func AccessUserAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define a struct to hold user information.
	var user models.User

	// Create a response object to handle errors.
	resp := err.ErrorResponse{
		Error: make(map[string]string),
	}

	// Extract email and password from the POST form data.
	err := r.ParseForm()
	if err != nil {
		log.Println("Error parsing form data:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))

	// Get the database connection from the pool.
	db := database.GetDB()

	// Query the database to retrieve user information.
	row := db.QueryRow("SELECT id, email, password FROM user WHERE email=?", email)
	err = row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		// Log the error and set appropriate response for invalid credentials.
		log.Println("Error executing SQL statement:", err)
		resp.Error["credentials"] = "Invalid credentials"

		if attempts, found := loginAttempts.Get(email); found {
			loginAttempts.Set(email, attempts.(int)+1, cache.DefaultExpiration)
		} else {
			loginAttempts.Set(email, 1, cache.DefaultExpiration)
		}

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Compare the provided password with the hashed password retrieved from the database.
	encErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if encErr != nil {
		// Set appropriate response for invalid password.
		resp.Error["password"] = "Invalid password"

		// Increment the login attempts counter for this user.
		if attempts, found := loginAttempts.Get(email); found {
			loginAttempts.Set(email, attempts.(int)+1, cache.DefaultExpiration)
		} else {
			loginAttempts.Set(email, 1, cache.DefaultExpiration)
		}

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// If there are errors, return the response.
	if len(resp.Error) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Println("Error encoding response:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	
	// If authentication is successful, remove the login attempts counter for this user.
	loginAttempts.Delete(email)

	// If authentication is successful, store user information in session and return success message.
	session := utils.GetSession(r)
	session.Values["id"] = strconv.Itoa(user.ID)
	session.Values["email"] = user.Email
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return success message if the user is created successfully.
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"message": "User logged in successfully"}`)); err != nil {
		log.Println("Error writing response:", err)
	}

}
