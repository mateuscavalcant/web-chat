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


func Chat(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	idInterface, _ := utils.AllSessions(r)
	id, _ := strconv.Atoi(idInterface.(string))

	// Obtendo o nome de usuário dos parâmetros da URL
	username := r.URL.Query().Get("username")

	db := database.GetDB()

	var messages []models.UserMessage

	var userID int
	err = db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Println("Failed to query user ID:", err)
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare(`
    SELECT user_message.message_id, user_message.id AS message_user_id, user_message.content,
           user.id AS user_id, user.username, user.name, user.icon
    FROM user_message
    JOIN user ON user.id = user_message.id
    WHERE (user_message.id = ? AND user_message.messageTo = ?) OR 
          (user_message.id = ? AND user_message.messageTo = ?)
    ORDER BY user_message.created_at ASC
	`)
	if err != nil {
		log.Println("Failed to prepare statement:", err)
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(id, userID, userID, id)
	if err != nil {
		log.Println("Failed to execute query:", err)
		http.Error(w, "Failed to execute query", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var message models.UserMessage
		var icon []byte

		err := rows.Scan(&message.MessageID, &message.MessageUserID, &message.Content, &message.UserID, &message.CreatedBy, &message.Name, &icon)
		if err != nil {
			log.Println("Failed to scan rows:", err)
			http.Error(w, "Failed to scan rows", http.StatusInternalServerError)
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
			CreatedBy:     message.CreatedBy,
			Name:          message.Name,
			IconBase64:    imageBase64,
		})
	}

	if err := conn.WriteJSON(messages); err != nil {
		log.Println("Failed to send messages over WebSocket:", err)
		return
	}
}


func CreateNewMessage(w http.ResponseWriter, r *http.Request) {
    var userMessage models.UserMessage
    var errResp err.ErrorResponse

    // Parse form data
    if err := r.ParseForm(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    username := r.FormValue("username")
    content := strings.TrimSpace(r.FormValue("content"))
    idInterface, _ := utils.AllSessions(r)
    if idInterface == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if content == "" {
        errResp.Error["content"] = "Values are missing!"
    }

    if len(errResp.Error) > 0 {
        errJSON, _ := json.Marshal(errResp)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        w.Write(errJSON)
        return
    }

    id, _ := strconv.Atoi(idInterface.(string))
    userMessage.Content = content

    db := database.GetDB()

    var userID int
    err := db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
    if err != nil {
        log.Println("Failed to query user ID", err)
        http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
        return
    }

    userMessage.MessageBy = id
    userMessage.MessageTo = userID

	stmt, err := db.Prepare("INSERT INTO user_message(content, messageBy, messageTo, userID, created_at) VALUES (?, ?, ?, ?, NOW())")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		http.Error(w, "Failed to prepare statement", http.StatusInternalServerError)
		return
	}

	defer stmt.Close()
	
	rs, err := stmt.Exec(userMessage.Content, userMessage.MessageBy, userMessage.MessageTo, id)
	if err != nil {
		log.Println("Error executing SQL statement:", err)
		http.Error(w, "Failed to execute statement", http.StatusInternalServerError)
		return
	}

    insertID, _ := rs.LastInsertId()

    resp := map[string]interface{}{
        "messageID": insertID,
        "message":   "Message sent successfully",
    }

    respJSON, _ := json.Marshal(resp)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(respJSON)

}