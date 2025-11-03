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
	
	// Validar configuración crítica
	if cfg.MongoURI == "" {
		log.Fatal("❌ MONGO_URI is required")
	}

	client := database.Connect(cfg.MongoURI)
	defer func() {
		if err := client.Disconnect(nil); err != nil {
			log.Println("❌ Error disconnecting from MongoDB:", err)
		}
	}()

	db := client.Database(cfg.MongoDB)

	router := gin.Default()
	routes.RegisterRoutes(router, db)

	// Usar el puerto de la configuración
	port := cfg.Port
	log.Printf("🚀 Server starting on port %s", port)
	
	// Escuchar en todas las interfaces
	if err := router.Run("0.0.0.0:" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}