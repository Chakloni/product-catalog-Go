package main

import (
	"log"
	"os"
	
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // valor por defecto para correr localmente
	}
	log.Println("🚀 Server running on port", port)
	router.Run(":" + port)

}
