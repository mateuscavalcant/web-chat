package handlers

import (
	_ "database/sql"
	"log"
	"net/http"
	"web-chat/api/utils"
	"web-chat/pkg/database"

	"github.com/gin-gonic/gin"
)

func FriendRequets(c *gin.Context) {
	// Retrieve the current user's ID from the session
	id, _ := utils.AllSessions(c)
	// Retrieve the username of the user to whom the friend request will be sent from the request body
	username := c.PostForm("username")

	db := database.GetDB()

	var userID int
	// Query the database to get the ID of the user to whom the friend request will be sent
	err := db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Println("Failed to query user ID", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user ID",
		})
		return
	}

	// Prepare an SQL statement to insert a new entry into the friend_request table
	stmt, err := db.Prepare("INSERT INTO friend_request(senderID, receiverID) VALUES(?, ?)")
	if err != nil {
		log.Println("Failed to prepare statement", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to prepare statement",
		})
		return
	}

	// Execute the SQL statement to insert the friend request
	_, err = stmt.Exec(id, userID)
	if err != nil {
		log.Println("Failed to execute query", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to execute query",
		})
		return
	}

	// Respond with a JSON message indicating the successful friend request action
	resp := map[string]interface{}{
		"mssg": "Friend request sent to ",
	}
	c.JSON(http.StatusOK, resp)
}

func ReceiveFriendRequest(c *gin.Context) {
	// Retrieve the current user's ID from the session
	userID, _ := utils.AllSessions(c)

	db := database.GetDB()

	// Query the database to get all friend requests sent to the current user
	rows, err := db.Query("SELECT id, senderID FROM friend_request WHERE receiverID = ?", userID)
	if err != nil {
		log.Println("Failed to query friend requests", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query friend requests",
		})
		return
	}
	defer rows.Close()

	// Slice to hold the friend requests
	friendRequests := []map[string]interface{}{}

	// Iterate over the rows returned by the query
	for rows.Next() {
		var requestID, senderID int
		err := rows.Scan(&requestID, &senderID)
		if err != nil {
			log.Println("Error scanning friend request rows", err)
			continue
		}

		// Retrieve sender information from the database
		var senderUsername string
		err = db.QueryRow("SELECT username FROM user WHERE id = ?", senderID).Scan(&senderUsername)
		if err != nil {
			log.Println("Failed to get sender username", err)
			continue
		}

		// Create a map to represent the friend request
		request := map[string]interface{}{
			"requestID":    requestID,
			"senderID":     senderID,
			"senderUsername": senderUsername,
		}

		// Append the friend request to the slice
		friendRequests = append(friendRequests, request)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over friend request rows", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error iterating over friend request rows",
		})
		return
	}

	// Respond with the list of friend requests
	c.JSON(http.StatusOK, gin.H{
		"friendRequests": friendRequests,
	})
}

func AcceptedFriendRequest(c *gin.Context) {
	// Retrieve the current user's ID from the session
	userID, _ := utils.AllSessions(c)
	// Retrieve the ID of the friend request to be accepted from the request body
	requestID := c.PostForm("requestID")

	db := database.GetDB()

	// Update the status of the friend request to 'accepted' in the database
	_, err := db.Exec("UPDATE friend_request SET status = 'accepted' WHERE id = ? AND receiverID = ?", requestID, userID)
	if err != nil {
		log.Println("Failed to update friend request status", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update friend request status",
		})
		return
	}

	// Add the accepted friend to the user's list of friends
	_, err = db.Exec("INSERT INTO user_friends(userID, friendID) VALUES(?, ?)", userID, requestID)
	if err != nil{
		log.Println("Failed to add friend", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add friend",
		})
		return
	}

	// Respond with a JSON message indicating the successful acceptance of the friend request
	resp := map[string]interface{}{
		"message": "Friend request accepted",
	}
	c.JSON(http.StatusOK, resp)
}
