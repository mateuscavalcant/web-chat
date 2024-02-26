package routes

import (
	"net/http"
	"web-chat/api/handlers"
	"web-chat/api/middlewares"

	"github.com/gorilla/mux"
)

func InitRoutes(r *mux.Router) {
	// Initialize the handlers
	r.HandleFunc("/signup", handlers.CreateUserAccount).Methods("POST")
	r.Handle("/login", middlewares.LimitLoginAttempts(http.HandlerFunc(handlers.AccessUserAccount)))
	r.HandleFunc("/chat/{username}", handlers.Chat).Methods("POST")
	r.HandleFunc("/create-message/{username}", handlers.CreateNewMessage).Methods("POST")
	r.HandleFunc("/websocket/{username}", handlers.WebSocketHandler)
}

