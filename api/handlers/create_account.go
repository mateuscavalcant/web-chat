package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"web-chat/pkg/database"
	"web-chat/pkg/models"
	"web-chat/pkg/models/err"
	"web-chat/pkg/validators"
)

// CreateUserAccount handles the creation of a new user account.
func CreateUserAccount(w http.ResponseWriter, r *http.Request) {
	// Define a struct to hold user information.
	var user models.User

	// Create a response object to handle errors.
	resp := err.ErrorResponse{
		Error: make(map[string]string),
	}

	// Extract form data from the request.
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing form:", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))
	confirmPassword := strings.TrimSpace(r.FormValue("confirm_password"))

	// Check if the email already exists in the database.
	existEmail, err := validators.ExistEmail(email)
	if err != nil {
		log.Println("Error checking email existence:", err)
		http.Error(w, "Failed to validate email", http.StatusInternalServerError)
		return
	}

	// Validate if all required fields are filled.
	if name == "" || email == "" || password == "" || confirmPassword == "" {
		resp.Error["missing"] = "Some values are missing!"
	}

	// Validate the length of the name.
	if len(name) < 4 || len(name) > 32 {
		resp.Error["name"] = "name should be between 4 and 32"
	}

	// Validate the email format.
	if validators.ValidateFormatEmail(email) != nil {
		resp.Error["email"] = "Invalid email format!"
	}

	// Check if the email already exists.
	if existEmail {
		resp.Error["email"] = "Email already exists!"
	}

	// Validate the length of the password.
	if len(password) < 8 || len(password) > 16 {
		resp.Error["password"] = "Passwords should be between 8 and 16"
	}

	// Check if passwords match.
	if password != confirmPassword {
		resp.Error["confirm_password"] = "Passwords don't match"
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

	// Populate the user struct with the provided data.
	user.Name = name
	user.Email = email
	user.Password = password

	// Get the database connection from the pool.
	db := database.GetDB()

	// Prepare SQL statement for user creation.
	query := "INSERT INTO user (name, email, password) VALUES (?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	// Execute the SQL statement to insert the user into the database.
	_, err = stmt.Exec(user.Name, user.Email, validators.Hash(user.Password))
	if err != nil {
		log.Println("Error executing SQL statement:", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Return success message if the user is created successfully.
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"message": "User created successfully"}`)); err != nil {
		log.Println("Error writing response:", err)
	}

	// Close the prepared statement to release resources.
	defer stmt.Close()
}
