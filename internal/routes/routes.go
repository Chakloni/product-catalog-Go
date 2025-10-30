package routes

import (
	"product-catalog/internal/handlers"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(router *gin.Engine, db *mongo.Database) {
	products := db.Collection("products")
	h := handlers.ProductHandler{Collection: products}

	v1 := router.Group("/v1")
	{
		v1.POST("/products", h.CreateProduct)
		v1.GET("/products", h.GetProducts)
		v1.GET("/products/:id", h.GetProductByID)
		v1.PATCH("/products/:id", h.UpdateProduct)
		v1.DELETE("/products/:id", h.DeleteProduct)
	}
}
