package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"product-catalog/internal/cache"
	"product-catalog/internal/config"
	"product-catalog/internal/handlers"
	"product-catalog/internal/middleware"
	"product-catalog/internal/repository"
)

func main() {
	// Configurar modo de Gin
	gin.SetMode(gin.ReleaseMode)

	// Inicializar base de datos
	if err := config.InitDB(); err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer config.Close()

	// Inicializar caché
	cache.Init(5 * time.Minute)
	log.Println("✅ Cache initialized successfully")

	// Inicializar repositorio y handler
	productRepo := repository.NewProductRepository(config.Collection)
	productHandler := handlers.NewProductHandler(productRepo)

	// Configurar router
	router := setupRouter(productHandler)

	// Puerto
	port := getEnv("PORT", "8080")

	// Servidor con graceful shutdown
	go func() {
		log.Printf("🚀 Server running on http://localhost:%s\n", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Esperar señal de interrupción
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	cache.Get().Clear()
	log.Println("✅ Server stopped gracefully")
}

func setupRouter(handler *handlers.ProductHandler) *gin.Engine {
	router := gin.New()

	// Middlewares globales
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Compresión GZIP
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Rate limiting (100 requests por minuto por IP)
	router.Use(middleware.RateLimiter(100))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := config.HealthCheck(); err != nil {
			c.JSON(500, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"status":     "healthy",
			"cache_size": cache.Get().Size(),
			"timestamp":  time.Now(),
		})
	})

	// API v1
	v1 := router.Group("/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", handler.CreateProduct)
			products.GET("", handler.ListProducts)
			products.GET("/:id", handler.GetProduct)
			products.PATCH("/:id", handler.UpdateProduct)
			products.DELETE("/:id", handler.DeleteProduct)
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "endpoint not found"})
	})

	return router
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}