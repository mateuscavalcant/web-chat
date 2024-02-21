package main

import (
	"os"
	"web-chat/api/routes"
	"web-chat/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


func main() {
	godotenv.Load()
	database.InitializeDB()
	r := gin.Default()

	r.Use(corsMiddleware())

	routes.InitRoutes(r.Group("/"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	r.Run(":" + port)
}


func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

	
