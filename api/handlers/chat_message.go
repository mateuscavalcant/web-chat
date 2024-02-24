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

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// userConnections mapeia IDs de usuário para conexões WebSocket.
var userConnections map[int64]*websocket.Conn

func init() {
    userConnections = make(map[int64]*websocket.Conn)
}

// Chat é um manipulador HTTP que lida com solicitações de chat.
func Chat(w http.ResponseWriter, r *http.Request) {
    // Extrair o ID do usuário da sessão
    idInterface, _ := utils.AllSessions(r)
    id, _ := strconv.Atoi(idInterface.(string))

    // Extrair o nome de usuário da solicitação
    vars := mux.Vars(r)
    username := vars["username"]

    // Obter o banco de dados
    db := database.GetDB()

    // Obter mensagens do banco de dados
    var messages []models.UserMessage

    // Obter o ID do usuário com base no nome de usuário
    var userID int
    err := db.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
    if err != nil {
        log.Println("Failed to query user ID:", err)
        http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
        return
    }

    // Preparar a consulta para buscar mensagens entre usuários
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

    // Executar a consulta SQL
    rows, err := stmt.Query(id, userID, userID, id)
    if err != nil {
        log.Println("Failed to execute query:", err)
        http.Error(w, "Failed to execute query", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    // Processar as linhas retornadas
    for rows.Next() {
        var message models.UserMessage
        var icon []byte

        err := rows.Scan(&message.MessageID, &message.MessageUserID, &message.Content, &message.UserID, &message.CreatedBy, &message.Name, &icon)
        if err != nil {
            log.Println("Failed to scan rows:", err)
            http.Error(w, "Failed to scan rows", http.StatusInternalServerError)
            return
        }

        // Codificar o ícone em base64, se existir
        var imageBase64 string
        if icon != nil {
            imageBase64 = base64.StdEncoding.EncodeToString(icon)
        }

        // Adicionar mensagem à lista de mensagens
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

    // Codificar as mensagens em JSON
    jsonData, err := json.Marshal(messages)
    if err != nil {
        log.Println("Failed to marshal JSON:", err)
        http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
        return
    }

    // Verificar se o usuário possui uma conexão WebSocket ativa
    conn, ok := userConnections[int64(id)]
    if !ok {
        log.Println("User does not have an active WebSocket connection")
    } else {
        // Enviar as mensagens via WebSocket
        err = conn.WriteJSON(messages)
        if err != nil {
            log.Println("Error sending messages via WebSocket:", err)
            // Tratar o erro de forma apropriada
        }
    }

    // Configurar o cabeçalho Content-Type para application/json
    w.Header().Set("Content-Type", "application/json")
    // Escrever os dados JSON na resposta
    w.Write(jsonData)
}

// HandleMessages lida com novas mensagens recebidas e as transmite para o destinatário.
func HandleMessages(ws *websocket.Conn) {
    defer ws.Close()

    for {
        var msg models.UserMessage
        err := ws.ReadJSON(&msg) // Use ReadJSON para ler mensagens JSON do WebSocket
        if err != nil {
            log.Println("Error receiving message:", err)
            return
        }

        // Verifique se o destinatário está conectado
        destConn, ok := userConnections[int64(msg.MessageTo)]
		if !ok {
			log.Println("Recipient is not connected")
			continue
}

        // Envie a mensagem para o destinatário
        err = destConn.WriteJSON(msg) // Use WriteJSON para enviar mensagens JSON via WebSocket
        if err != nil {
            log.Println("Error sending message:", err)
            continue
        }
    }
}

// WebSocketHandler é um manipulador HTTP para a rota WebSocket.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
    // Upgrade para WebSocket
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if err != nil {
        log.Println("Error upgrading to WebSocket:", err)
        return
    }
    defer ws.Close()

    // Registre a conexão com o usuário
    idInterface, _ := utils.AllSessions(r)
    id, _ := strconv.Atoi(idInterface.(string))
    userConnections[int64(id)] = ws // Armazene o ponteiro ws, que é do tipo *websocket.Conn

    // Aguardar mensagens do usuário
    HandleMessages(ws)
}



func CreateNewMessage(w http.ResponseWriter, r *http.Request) {
    var userMessage models.UserMessage
    var errResp err.ErrorResponse

    // Parse form data
    if err := r.ParseForm(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    username := vars["username"]
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

	stmt, err := db.Prepare("INSERT INTO user_message(content, messageBy, messageTo, id, created_at) VALUES (?, ?, ?, ?, NOW())")
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