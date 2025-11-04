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