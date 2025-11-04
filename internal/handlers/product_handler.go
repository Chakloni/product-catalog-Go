package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"product-catalog/internal/models"
	"product-catalog/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

type ProductHandler struct {
	Repo *repository.ProductRepository
}

func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{Repo: repo}
}

// GetAllProducts obtiene la lista de productos con filtros y paginación
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	category := c.Query("category")
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	summary := c.DefaultQuery("summary", "false") == "true"

	products, total, err := h.Repo.FindAll(
		c.Request.Context(),
		page,
		pageSize,
		category,
		sortBy,
		sortOrder,
		summary,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"products": products,
	})
}

// GetProductByID obtiene un producto por su ID
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	product, err := h.Repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// CreateProduct crea un nuevo producto
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repo.Create(c.Request.Context(), &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct actualiza un producto
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var update models.ProductUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateMap := bson.M{}
	if update.Name != nil {
		updateMap["name"] = *update.Name
	}
	if update.Description != nil {
		updateMap["description"] = *update.Description
	}
	if update.Category != nil {
		updateMap["category"] = *update.Category
	}
	if update.PriceCents != nil {
		updateMap["price_cents"] = *update.PriceCents
	}
	if update.Currency != nil {
		updateMap["currency"] = *update.Currency
	}
	if update.Stock != nil {
		updateMap["stock"] = *update.Stock
	}
	if update.Images != nil {
		updateMap["images"] = update.Images
	}
	if update.Attributes != nil {
		updateMap["attributes"] = update.Attributes
	}
	if update.IsActive != nil {
		updateMap["is_active"] = *update.IsActive
	}

	if len(updateMap) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields to update"})
		return
	}

	if err := h.Repo.Update(c.Request.Context(), id, updateMap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}

// DeleteProduct realiza un borrado lógico
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	if err := h.Repo.SoftDelete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}
