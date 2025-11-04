package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"

    "product-catalog/internal/models"
    "product-catalog/internal/repository"
)

type ProductHandler struct {
    Repo *repository.ProductRepository
}

func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
    return &ProductHandler{Repo: repo}
}

// POST /v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var product models.Product
    if err := c.ShouldBindJSON(&product); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    product.CreatedAt = time.Now()
    product.UpdatedAt = time.Now()
    product.IsActive = true

    if err := h.Repo.Create(&product); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, product)
}

// GET /v1/products
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
    products, err := h.Repo.GetAll()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, products)
}

// GET /v1/products/:id
func (h *ProductHandler) GetProductByID(c *gin.Context) {
    id := c.Param("id")

    product, err := h.Repo.GetByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
        return
    }

    c.JSON(http.StatusOK, product)
}

// PATCH /v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    id := c.Param("id")
    var updateData models.Product

    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    update := bson.M{}
    if updateData.Name != "" {
        update["name"] = updateData.Name
    }
    if updateData.Description != "" {
        update["description"] = updateData.Description
    }
    if updateData.Price > 0 {
        update["price"] = updateData.Price
    }
    if updateData.Category != "" {
        update["category"] = updateData.Category
    }
    if updateData.Stock >= 0 {
        update["stock"] = updateData.Stock
    }

    update["updated_at"] = time.Now()

    if err := h.Repo.Update(id, update); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}

// DELETE /v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    id := c.Param("id")

    if err := h.Repo.SoftDelete(id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}
