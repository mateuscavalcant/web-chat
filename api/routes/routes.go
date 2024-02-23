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
	r.Handle("/login", middlewares.LimitLoginAttempts(http.HandlerFunc(handlers.AccessUserAccount))).Methods("POST")
	r.Handle("/chat/{username}", middlewares.AuthMiddleware(http.HandlerFunc(handlers.Chat))).Methods("POST")
	r.Handle("/create-message/{username}", middlewares.AuthMiddleware(http.HandlerFunc(handlers.CreateNewMessage))).Methods("POST")
}
