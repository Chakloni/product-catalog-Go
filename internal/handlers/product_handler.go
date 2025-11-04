package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "path/to/your/project/internal/services"
    "path/to/your/project/internal/models"
)

type ProductHandler struct {
    ProductService services.ProductService
}

func NewProductHandler(ps services.ProductService) *ProductHandler {
    return &ProductHandler{
        ProductService: ps,
    }
}

// CreateProduct handles POST /v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req models.ProductCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // mapear a modelo interno
    prod := models.Product{
        SKU:         req.SKU,
        Name:        req.Name,
        Description: req.Description,
        Category:    req.Category,
        PriceCents:  req.PriceCents,
        Currency:    req.Currency,
        Stock:       req.Stock,
        Images:      req.Images,
        Attributes:  req.Attributes,
        IsActive:    true,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    created, err := h.ProductService.Create(c.Request.Context(), &prod)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, created)
}

// ListProducts handles GET /v1/products
func (h *ProductHandler) ListProducts(c *gin.Context) {
    // Opciones de paginación, filtro, ordenamiento podrían venir via query params
    page := c.DefaultQuery("page", "1")
    size := c.DefaultQuery("size", "10")
    // parsear page/size a int, etc. Omitido aquí por brevedad.

    filter := make(map[string]interface{})
    // Ejemplo: filtrar por category, price range, etc
    if cat := c.Query("category"); cat != "" {
        filter["category"] = cat
    }

    // Solo activos
    filter["is_active"] = true

    // llamar al servicio
    result, err := h.ProductService.List(c.Request.Context(), filter, page, size)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, result)
}

// GetProduct handles GET /v1/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
    idParam := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
        return
    }

    prod, err := h.ProductService.GetByID(c.Request.Context(), objID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
        return
    }

    c.JSON(http.StatusOK, prod)
}

// UpdateProduct handles PATCH /v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    idParam := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
        return
    }

    var req models.ProductUpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // crear un map de campos a actualizar
    updates := make(map[string]interface{})
    if req.Name != nil {
        updates["name"] = *req.Name
    }
    if req.Description != nil {
        updates["description"] = *req.Description
    }
    if req.Category != nil {
        updates["category"] = *req.Category
    }
    if req.PriceCents != nil {
        updates["price_cents"] = *req.PriceCents
    }
    if req.Currency != nil {
        updates["currency"] = *req.Currency
    }
    if req.Stock != nil {
        updates["stock"] = *req.Stock
    }
    if req.Images != nil {
        updates["images"] = *req.Images
    }
    if req.Attributes != nil {
        updates["attributes"] = *req.Attributes
    }

    updates["updated_at"] = time.Now()

    updated, err := h.ProductService.Update(c.Request.Context(), objID, updates)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, updated)
}

// DeleteProduct handles DELETE /v1/products/:id — soft delete
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    idParam := c.Param("id")
    objID, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
        return
    }

    err = h.ProductService.SoftDelete(c.Request.Context(), objID, time.Now())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusNoContent, gin.H{})
}
