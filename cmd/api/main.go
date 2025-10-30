package main

import (
	"log"
	"product-catalog/internal/config"
	"product-catalog/internal/database"
	"product-catalog/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	client := database.Connect(cfg.MongoURI)
	db := client.Database(cfg.MongoDB)

	router := gin.Default()
	routes.RegisterRoutes(router, db)

	log.Println("🚀 Server running on port", cfg.Port)
	router.Run(":" + cfg.Port)
}
