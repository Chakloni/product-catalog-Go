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

type ProductHandler struct {
	Collection *mongo.Collection
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
	product.IsDeleted = false

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.Collection.InsertOne(ctx, product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not insert product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GET /v1/products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filters
	filter := bson.M{
		"$or": []bson.M{
			{"is_deleted": bson.M{"$exists": false}},
			{"is_deleted": false},
		},
	}

	if q := c.Query("q"); q != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": q, "$options": "i"}},
			{"description": bson.M{"$regex": q, "$options": "i"}},
			{"category": bson.M{"$regex": q, "$options": "i"}},
		}
	}
	if cat := c.Query("category"); cat != "" {
		filter["category"] = cat
	}
	if active := c.Query("active"); active != "" {
		if active == "true" {
			filter["is_active"] = true
		} else if active == "false" {
			filter["is_active"] = false
		}
	}

	// Price filters
	minPrice, _ := strconv.ParseInt(c.Query("min_price"), 10, 64)
	maxPrice, _ := strconv.ParseInt(c.Query("max_price"), 10, 64)
	priceFilter := bson.M{}
	if minPrice > 0 {
		priceFilter["$gte"] = minPrice
	}
	if maxPrice > 0 {
		priceFilter["$lte"] = maxPrice
	}
	if len(priceFilter) > 0 {
		filter["price_cents"] = priceFilter
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	skip := (page - 1) * pageSize

	// Sorting
	sort := bson.D{}
	sortQuery := c.DefaultQuery("sort", "name:asc")
	for _, part := range strings.Split(sortQuery, ",") {
		fields := strings.Split(part, ":")
		field := fields[0]
		order := 1
		if len(fields) > 1 && fields[1] == "desc" {
			order = -1
		}
		sort = append(sort, bson.E{Key: field, Value: order})
	}

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(sort)

	cursor, err := h.Collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch products"})
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error decoding products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"products":  products,
	})
}

// GET /v1/products/:id
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idParam := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product models.Product
	err = h.Collection.FindOne(ctx, bson.M{"_id": objID, "is_deleted": false}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// PATCH /v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idParam := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var update bson.M
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	update["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateResult, err := h.Collection.UpdateOne(ctx,
		bson.M{"_id": objID, "is_deleted": false},
		bson.M{"$set": update})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update product"})
		return
	}

	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

// DELETE /v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	result, err := h.Collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete product"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product marked as deleted"})
}

