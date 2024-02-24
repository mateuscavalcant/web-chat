package routes

import (
	"net/http"
	"web-chat/api/handlers"
	"web-chat/api/middlewares"
	"web-chat/api/views"

	"github.com/gorilla/mux"
)

func InitRoutes(r *mux.Router) {
	// Initialize the handlers
	r.HandleFunc("/signup", handlers.CreateUserAccount).Methods("POST")
	r.HandleFunc("/login", views.RenderLoginTemplate).Methods("GET")
	r.Handle("/login", middlewares.LimitLoginAttempts(http.HandlerFunc(handlers.AccessUserAccount)))
	r.HandleFunc("/chat/{username}", handlers.Chat).Methods("POST")
	r.HandleFunc("/create-message/{username}", handlers.CreateNewMessage).Methods("POST")
	r.HandleFunc("/websocket/{username}", handlers.WebSocketHandler)
	r.Handle("/home", middlewares.AuthMiddleware(http.HandlerFunc(views.RenderHomeTemplate))).Methods("GET")
	r.Handle("/chat/{username}", middlewares.AuthMiddleware(http.HandlerFunc(views.RenderChatTemplate))).Methods("GET")
}

