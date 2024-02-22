package handlers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"web-chat/api/utils"
	"web-chat/pkg/database"
	"web-chat/pkg/models"
	"web-chat/pkg/models/err"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	clients = make(map[*websocket.Conn]*Client)
)

func Chat(c *gin.Context) {

	idInterface, _ := utils.AllSessions(c)
	id, _ := strconv.Atoi(idInterface.(string))

    username := c.Param("username") // Obtenha o nome de usuário dos parâmetros da rota

    db := database.GetDB()

    var messages []models.UserMessage

    // Obtendo o ID do usuário com base no nome de usuário fornecido
    var userID int
    err := db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
    if err != nil {
        log.Println("Failed to query user ID:", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get user ID",
        })
        return
    }

    query := `
        SELECT user_message.message_id, user_message.id AS message_user_id, user_message.content,
               user.id AS user_id, user.username, user.name, user.icon
        FROM user_message
        JOIN user ON user.id = user_message.id
        WHERE (user_message.id = ? AND user_message.messageTo = ?) OR 
              (user_message.id = ? AND user_message.messageTo = ?)
        ORDER BY user_message.created_at ASC
    `

    rows, err := db.Query(query, id, userID, userID, id)
    if err != nil {
        log.Println("Failed to execute query:", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to execute query",
        })
        return
    }
    defer rows.Close()

    for rows.Next() {
        var message models.UserMessage
        var icon []byte

        err := rows.Scan(&message.MessageID, &message.MessageUserID, &message.Content, &message.UserID, &message.MessageBy, &message.Name, &icon)
        if err != nil {
            log.Println("Failed to scan rows:", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to scan rows",
            })
            return
        }

        var imageBase64 string
        if icon != nil {
            imageBase64 = base64.StdEncoding.EncodeToString(icon)
        }

        messages = append(messages, models.UserMessage{
            MessageID:     message.MessageID,
            MessageUserID: message.MessageUserID,
            Content:       message.Content,
            UserID:        message.UserID,
            MessageBy:     message.MessageBy,
            Name:          message.Name,
            IconBase64:    imageBase64,
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "messages": messages,
    })
}

func CreateNewMessage(c *gin.Context) {
    var userMessage models.UserMessage
    errresp := err.ErrorResponse{
        Error: make(map[string]string),
    }

    username := c.PostForm("username")
    content := strings.TrimSpace(c.PostForm("content"))
    idInterface, _ := utils.AllSessions(c)
    if idInterface == nil {
        // If the user is not logged in, return an authentication error
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Unauthorized",
        })
        return
    }

    if content == "" {
        errresp.Error["content"] = "Values are missing!"
    }

    if len(errresp.Error) > 0 {
        c.JSON(400, errresp)
        return
    }

    id, _ := strconv.Atoi(idInterface.(string))
    userMessage.Content = content

    db := database.GetDB()

    var usernameSession string
    err := db.QueryRow("SELECT username FROM user WHERE id = ?", id).Scan(&usernameSession)
    if err != nil {
        log.Println("Error querying username:", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to query username",
        })
        return
    }

    var userID int
    // Query the database to get the ID of the user to be followed
    errUsername := db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
    if err != nil {
        log.Println("Failed to query user ID", errUsername)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get user ID",
        })
        return
    }

    userMessage.MessageBy = usernameSession
    userMessage.MessageTo = userID

    stmt, err := db.Prepare("INSERT INTO user_message(content, messageBy, messageTo, userID, created_at) VALUES (?, ?, ?, ?, NOW())")

    if err != nil {
        log.Println("Error preparing SQL statement:", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to prepare statement",
        })
        return
    }

    rs, err := stmt.Exec(userMessage.Content, userMessage.MessageBy, userMessage.MessageTo, id)
    if err != nil {
        log.Println("Error executing SQL statement:", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to execute statement",
        })
        return
    }

    insertID, _ := rs.LastInsertId()

    // Prepare the message to send via WebSocket
    messageToSend := map[string]interface{}{
        "messageID": insertID,
        "mssg":      "message Created!!",
    }

    // Convert the message to JSON
    jsonMessage, err := json.Marshal(messageToSend)
    if err != nil {
        log.Println("Error marshaling JSON:", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to marshal JSON",
        })
        return
    }

    // Send the message to all WebSocket clients
    for _, client := range clients {
        if err := client.conn.WriteMessage(websocket.TextMessage, jsonMessage); err != nil {
            log.Println("Error writing WebSocket message:", err)
            // If there's an error, continue sending to other clients
            continue
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Message sent successfully",
    })
}