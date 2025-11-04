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

	// Buscar en DB
	product, err := h.repo.FindByID(c.Request.Context(), productID)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get product"})
		return
	}

	// Guardar en caché
	h.cache.Set(cacheKey, product, 5*time.Minute)
	c.JSON(http.StatusOK, product)
}

// ListProducts lista productos con paginación y filtros (con caché)
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	summary := c.DefaultQuery("summary", "false") == "true"

	cacheKey := fmt.Sprintf(
		"products:list:p%d_s%d_cat:%s_sort:%s_%s_sum:%v",
		page, pageSize, category, sortBy, sortOrder, summary,
	)

	// Buscar en caché
	if cached, found := h.cache.GetValue(cacheKey); found {
		c.JSON(http.StatusOK, cached)
		return
	}

	// Buscar en base de datos
	products, total, err := h.repo.FindAll(
		c.Request.Context(),
		page,
		pageSize,
		category,
		sortBy,
		sortOrder,
		summary,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}

	response := gin.H{
		"data":       products,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": func() int64 {
			if pageSize == 0 {
				return 1
			}
			tp := total / int64(pageSize)
			if total%int64(pageSize) != 0 {
				tp++
			}
			return tp
		}(),
	}

	// Guardar en caché
	h.cache.Set(cacheKey, response, 2*time.Minute)
	c.JSON(http.StatusOK, response)
}

// UpdateProduct actualiza parcialmente un producto
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
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

	if err := h.repo.Update(c.Request.Context(), productID, updateMap); err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	// Invalidar caché relacionado
	h.cache.Delete(fmt.Sprintf("product:%s", productID))
	h.cache.DeleteByPrefix("products:list:")

	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

// DeleteProduct realiza un borrado lógico
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

	// Invalidar caché relacionado
	h.cache.Delete(fmt.Sprintf("product:%s", productID))
	h.cache.DeleteByPrefix("products:list:")

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}
