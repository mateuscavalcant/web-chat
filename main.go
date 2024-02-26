package main

import (
	"net/http"
	"web-chat/api/routes"
	"web-chat/pkg/database"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.InitializeDB()
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	routes.InitRoutes(router)
	
	http.ListenAndServe(":8080", router)

}