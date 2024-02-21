package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"web-chat/pkg/database"
	"web-chat/pkg/models"
	"web-chat/pkg/models/err"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func DeleteUserAccount(c *gin.Context) {
	// Define a struct to hold user information.
	var user models.User

	// Extract data from the JSON request body.
    var data map[string]string
    if err := c.ShouldBindJSON(&data); err != nil {
        log.Println("Error binding JSON:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
        return
    }

    email := strings.TrimSpace(data["email"])
    password := strings.TrimSpace(data["password"])
    confirmPassword := strings.TrimSpace(data["confirm_password"])

	// Create a response object to handle errors.
	resp := err.ErrorResponse{
		Error: make(map[string]string),
	}

	    // Get the database connection from the pool.
		db := database.GetDB()

		// Query the database to retrieve user information.
		log.Println("Querying user with email:", email)
		row := db.QueryRow("SELECT id, email, password FROM user WHERE email=?", email)
		err := row.Scan(&user.ID, &user.Email, &user.Password)
        if err != nil {
			if err == sql.ErrNoRows {
                // Log the error and set appropriate response for invalid credentials.
                log.Println("No user found with email:", email)
                resp.Error["email"] = "Invalid credentials"
            } else {
                log.Println("Error executing SQL statement:", err)
                resp.Error["email"] = "Database error"
            }
            c.JSON(400, resp)
            return
        }
	
		// Compare the provided password with the hashed password retrieved from the database.
		encErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if encErr != nil || password != confirmPassword {
			// Set appropriate response for invalid password.
			resp.Error["password"] = "Invalid password"
		}
	
		// If there are errors, return the response.
		if len(resp.Error) > 0 {
			c.JSON(400, resp)
			return
		}
	
		// If authentication is successful, delete user.
		stmt, errDB := db.Prepare("DELETE FROM user WHERE id=?")
		if errDB != nil {
			log.Println("Error preparing SQL statement:", errDB)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}
		defer stmt.Close()
	
		// Use the user ID obtained from the database to delete the user.
		_, errDB = stmt.Exec(user.ID)
		if errDB != nil {
			log.Println("Error deleting user:", errDB)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}
	
		log.Println("User deleted")
		c.JSON(200, gin.H{"message": "User deleted successfully"})
	}