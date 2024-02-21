package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"web-chat/api/utils"
	"web-chat/pkg/database"
	"web-chat/pkg/models/err"

	"github.com/gin-gonic/gin"
)

// UpdateUserAccount handles the update of user account information.
func UpdateUserAccount(c *gin.Context) {
	// Retrieve user ID from session.
	_, idInterface := utils.AllSessions(c)
	id, _ := strconv.Atoi(idInterface.(string))

	// Read and process file data if present.
	var fileBytes []byte
	file, _, errFile := c.Request.FormFile("icon")
	if errFile != nil && errFile != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error getting image from the form"})
		return
	} else if errFile == nil {
		defer file.Close()

		// Read file data into byte slice.
		fileBytes, errFile = ioutil.ReadAll(file)
		if errFile != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading the image"})
			return
		}
	}

	// Extract form data.
	name := strings.TrimSpace(c.PostForm("name"))
	bio := strings.TrimSpace(c.PostForm("bio"))

	// Create a response object to handle errors.
	resp := err.ErrorResponse{
		Error: make(map[string]string),
	}

	// Validate name length.
	if len(name) < 1 || len(name) > 70 {
		resp.Error["name"] = "Name should be between 1 and 70"
	}

	// Validate bio length.
	if len(bio) > 150 {
		resp.Error["bio"] = "Bio should be between 0 and 150"
	}

	// If there are errors, return the response.
	if len(resp.Error) > 0 {
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Get the database connection.
	db := database.GetDB()
	defer db.Close() // Close the database connection at the end of the function.

	stmt, errDB := db.Prepare("UPDATE user SET name=?, bio=? WHERE id=?")
	if errDB != nil {
		log.Println("Error preparing SQL statement:", errDB)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	defer stmt.Close()

	if fileBytes != nil {
		stmt, errDB = db.Prepare("UPDATE user SET name=?, bio=?, icon=? WHERE id=?")
		if errDB != nil {
			log.Println("Error preparing SQL statement:", errDB)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		defer stmt.Close()

		_, errDB = stmt.Exec(name, bio, fileBytes, id)
	} else {
		stmt, errDB = db.Prepare("UPDATE user SET name=?, bio=? WHERE id=?")
		if errDB != nil {
			log.Println("Error preparing SQL statement:", errDB)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
		defer stmt.Close()

		_, errDB = stmt.Exec(name, bio, id)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})

}