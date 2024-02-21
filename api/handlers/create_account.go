package handlers

import (
	_ "database/sql"
	"log"
	"strings"
	"web-chat/pkg/database"
	"web-chat/pkg/models"
	"web-chat/pkg/models/err"
	"web-chat/pkg/validators"

	"github.com/gin-gonic/gin"
)

// CreateUserAccount handles the creation of a new user account.
func CreateUserAccount(c *gin.Context) {
	// Define a struct to hold user information.
	var user models.User

	// Create a response object to handle errors.
	resp := err.ErrorResponse{
		Error: make(map[string]string),
	}

	// Extract form data from the request.
	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(c.PostForm("email"))
	password := strings.TrimSpace(c.PostForm("password"))
	confirmPassword := strings.TrimSpace(c.PostForm("confirm_password"))

	// Check if the email already exists in the database.
	existEmail, err := validators.ExistEmail(email)
	if err != nil {
		log.Println("Error checking email existence:", err)
		c.JSON(500, gin.H{"error": "Failed to validate email"})
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
		c.JSON(400, resp)
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
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	// Return success message if the user is created successfully.
	c.JSON(200, gin.H{"message": "User created successfully"})

	// Close the prepared statement to release resources.
	defer stmt.Close()
}
