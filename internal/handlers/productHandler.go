package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"product-catalog/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultPage     = 1
	defaultPageSize = 10
	defaultTimeout  = 5 * time.Second
	queryTimeout    = 10 * time.Second
)

type ProductHandler struct {
	Collection *mongo.Collection
}

// Estructuras para respuestas
type ProductListResponse struct {
	Page      int              `json:"page"`
	PageSize  int              `json:"page_size"`
	Total     int64            `json:"total,omitempty"`
	Products  []models.Product `json:"products"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

// POST /v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Validaciones básicas
	if err := h.validateProduct(&product); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Inicializar campos del sistema
	product.ID = primitive.NewObjectID()
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now
	product.IsDeleted = false

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if _, err := h.Collection.InsertOne(ctx, product); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not insert product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GET /v1/products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	filter := h.buildFilter(c)
	page, pageSize := h.getPaginationParams(c)
	sortOptions := h.buildSortOptions(c)

	opts := options.Find().
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize)).
		SetSort(sortOptions)

	cursor, err := h.Collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not fetch products"})
		return
	}
	defer cursor.Close(ctx)

	products := make([]models.Product, 0)
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "error decoding products"})
		return
	}

	// Opcional: obtener el total de documentos para paginación más completa
	// total, _ := h.Collection.CountDocuments(ctx, filter)

	c.JSON(http.StatusOK, ProductListResponse{
		Page:     page,
		PageSize: pageSize,
		Products: products,
		// Total:    total,
	})
}

// GET /v1/products/:id
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	objID, err := h.parseObjectID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid product ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var product models.Product
	filter := bson.M{"_id": objID, "is_deleted": false}
	
	if err := h.Collection.FindOne(ctx, filter).Decode(&product); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "error fetching product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// PATCH /v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	objID, err := h.parseObjectID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid product ID"})
		return
	}

	var update bson.M
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Prevenir actualización de campos del sistema
	h.sanitizeUpdate(update)
	update["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	filter := bson.M{"_id": objID, "is_deleted": false}
	result, err := h.Collection.UpdateOne(ctx, filter, bson.M{"$set": update})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not update product"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "product not found"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "product updated successfully"})
}

// DELETE /v1/products/:id (soft delete)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	objID, err := h.parseObjectID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid product ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	result, err := h.Collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "could not delete product"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "product not found"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "product deleted successfully"})
}

// --- Métodos auxiliares ---

// parseObjectID convierte un string a ObjectID
func (h *ProductHandler) parseObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

// buildFilter construye el filtro de MongoDB basado en query params
func (h *ProductHandler) buildFilter(c *gin.Context) bson.M {
	filter := bson.M{"is_deleted": false}

	// Búsqueda de texto
	if q := c.Query("q"); q != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": q, "$options": "i"}},
			{"description": bson.M{"$regex": q, "$options": "i"}},
			{"category": bson.M{"$regex": q, "$options": "i"}},
		}
	}

	// Filtro por categoría
	if cat := c.Query("category"); cat != "" {
		filter["category"] = cat
	}

	// Filtro por estado activo
	if active := c.Query("active"); active != "" {
		filter["is_active"] = active == "true"
	}

	// Filtros de precio
	h.addPriceFilter(filter, c)

	return filter
}

// addPriceFilter agrega filtros de precio al filtro principal
func (h *ProductHandler) addPriceFilter(filter bson.M, c *gin.Context) {
	priceFilter := bson.M{}

	if minPrice, err := strconv.ParseInt(c.Query("min_price"), 10, 64); err == nil && minPrice > 0 {
		priceFilter["$gte"] = minPrice
	}

	if maxPrice, err := strconv.ParseInt(c.Query("max_price"), 10, 64); err == nil && maxPrice > 0 {
		priceFilter["$lte"] = maxPrice
	}

	if len(priceFilter) > 0 {
		filter["price_cents"] = priceFilter
	}
}

// getPaginationParams obtiene y valida los parámetros de paginación
func (h *ProductHandler) getPaginationParams(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))

	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = defaultPageSize
	}

	return page, pageSize
}

// buildSortOptions construye las opciones de ordenamiento
func (h *ProductHandler) buildSortOptions(c *gin.Context) bson.D {
	sortQuery := c.DefaultQuery("sort", "name:asc")
	sort := bson.D{}

	for _, part := range strings.Split(sortQuery, ",") {
		fields := strings.Split(strings.TrimSpace(part), ":")
		if len(fields) == 0 {
			continue
		}

		field := fields[0]
		order := 1
		if len(fields) > 1 && fields[1] == "desc" {
			order = -1
		}

		sort = append(sort, bson.E{Key: field, Value: order})
	}

	return sort
}

// sanitizeUpdate elimina campos que no deben ser actualizados por el usuario
func (h *ProductHandler) sanitizeUpdate(update bson.M) {
	protectedFields := []string{"_id", "id", "created_at", "is_deleted"}
	for _, field := range protectedFields {
		delete(update, field)
	}
}

// validateProduct valida los campos requeridos del producto
func (h *ProductHandler) validateProduct(p *models.Product) error {
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if p.SKU == "" {
		return &ValidationError{Field: "sku", Message: "SKU is required"}
	}
	if p.PriceCents < 0 {
		return &ValidationError{Field: "price_cents", Message: "price cannot be negative"}
	}
	if p.Stock < 0 {
		return &ValidationError{Field: "stock", Message: "stock cannot be negative"}
	}
	return nil
}

// ValidationError representa un error de validación
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}