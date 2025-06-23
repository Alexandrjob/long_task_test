package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	manager := NewTaskManager()
	handler := NewTaskHandler(manager)

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	gin.ErrorLogger()

	apiV1 := router.Group("/api/v1")
	{
		task := apiV1.Group("/task")
		task.POST("/create", handler.Create)
		task.DELETE("/:id/delete", handler.Delete)
		task.GET("/:id/info", handler.GetInfo)
		task.GET("/:id/result", handler.GetResult)
	}

	router.Run(":8080")
}
