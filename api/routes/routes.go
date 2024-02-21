package routes

import (
	"web-chat/api/handlers"
	"web-chat/api/middlewares"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.RouterGroup) {
	r.POST("/signup", handlers.CreateUserAccount)
	r.POST("/login", middlewares.LimitLoginAttempts(),handlers.AccessUserAccount)
	r.PUT("/update-profile", handlers.UpdateUserAccount)
	r.DELETE("/delete-account", handlers.DeleteUserAccount)
}