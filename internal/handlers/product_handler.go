package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"

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

    product.ID = primitive.NewObjectID()
    product.CreatedAt = time.Now()
    product.UpdatedAt = time.Now()
    product.IsActive = true

    if err := h.Repo.Create(c, &product); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, product)
}

// GET /v1/products
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
    products, err := h.Repo.GetAll(c, true) // true → solo activos
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, products)
}

// GET /v1/products/:id
func (h *ProductHandler) GetProductByID(c *gin.Context) {
    idParam := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
        return
    }

    product, err := h.Repo.GetByID(c, objID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
        return
    }

    c.JSON(http.StatusOK, product)
}

// PATCH /v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    idParam := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
        return
    }

    var updateData models.Product
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    updateData.UpdatedAt = time.Now()

    if err := h.Repo.Update(c, objID, updateData); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}

// DELETE /v1/products/:id
// Borrado lógico: cambia is_active = false
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    idParam := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
        return
    }

    if err := h.Repo.SoftDelete(c, objID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}
