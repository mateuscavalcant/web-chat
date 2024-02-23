package main

import (
	"net/http"
	"web-chat/api/routes"
	"web-chat/pkg/database"

	"github.com/gorilla/mux"
)

func main() {
	database.InitializeDB()
	router := mux.NewRouter()
	routes.InitRoutes(router)

	http.ListenAndServe(":8765", router)

}