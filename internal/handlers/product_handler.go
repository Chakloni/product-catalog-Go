package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"product-catalog/internal/cache"
	"product-catalog/internal/models"
	"product-catalog/internal/repository"
)

type ProductHandler struct {
	repo  *repository.ProductRepository
	cache *cache.Cache
}

func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		repo:  repo,
		cache: cache.Get(),
	}
}

// CreateProduct crea un nuevo producto
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(c.Request.Context(), &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}

	// Invalidar caché de listados
	h.cache.DeleteByPrefix("products:list:")

	c.JSON(http.StatusCreated, product)
}

// GetProduct obtiene un producto por ID (con caché)
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	cacheKey := fmt.Sprintf("product:%s", productID)

	// Intentar obtener del caché
	if cachedProduct, found := h.cache.GetValue(cacheKey); found {
		c.JSON(http.StatusOK, cachedProduct)
		return
	}

	// Si no está en caché, buscar en DB
	product, err := h.repo.FindByID(c.Request.Context(), productID)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get product"})
		return
	}

	// Guardar en caché por 5 minutos
	h.cache.Set(cacheKey, product, 5*time.Minute)

	c.JSON(http.StatusOK, product)
}

// ListProducts lista productos con paginación y filtros (con caché)
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// Parsear parámetros
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	summary := c.DefaultQuery("summary", "false") == "true"

	// Cache key basado en parámetros
	cacheKey := fmt.Sprintf("products:list:%d:%d:%s:%s:%s:%v",
		page, pageSize, category, sortBy, sortOrder, summary)

	// Intentar obtener del caché
	type CachedResponse struct {
		Products []*models.Product `json:"products"`
		Total    int64             `json:"total"`
		Page     int               `json:"page"`
		PageSize int               `json:"page_size"`
	}

	var response CachedResponse
	if found, err := h.cache.Unmarshal(cacheKey, &response); err == nil && found {
		c.JSON(http.StatusOK, response)
		return
	}

	// Si no está en caché, buscar en DB
	products, total, err := h.repo.FindAll(c.Request.Context(), page, pageSize, category, sortBy, sortOrder, summary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}

	response = CachedResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	// Guardar en caché por 2 minutos
	h.cache.Marshal(cacheKey, response, 2*time.Minute)

	c.JSON(http.StatusOK, response)
}

// UpdateProduct actualiza un producto
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertir a bson.M
	update := bson.M{}
	for key, value := range updateData {
		if key != "_id" && key != "created_at" && key != "is_deleted" {
			update[key] = value
		}
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields to update"})
		return
	}

	if err := h.repo.Update(c.Request.Context(), productID, update); err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	// Invalidar caché
	h.cache.Delete(fmt.Sprintf("product:%s", productID))
	h.cache.DeleteByPrefix("products:list:")

	c.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}

// DeleteProduct elimina (soft delete) un producto
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	if err := h.repo.SoftDelete(c.Request.Context(), productID); err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}

	// Invalidar caché
	h.cache.Delete(fmt.Sprintf("product:%s", productID))
	h.cache.DeleteByPrefix("products:list:")

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}