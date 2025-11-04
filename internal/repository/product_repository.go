package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"product-catalog/internal/models"
)

type ProductRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(collection *mongo.Collection) *ProductRepository {
	return &ProductRepository{
		collection: collection,
	}
}

// Create crea un nuevo producto
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.IsDeleted = false

	_, err := r.collection.InsertOne(ctx, product)
	return err
}

// FindByID obtiene un producto por ID
func (r *ProductRepository) FindByID(ctx context.Context, id string) (*models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID")
	}

	var product models.Product
	filter := bson.M{
		"_id":        objID,
		"is_deleted": false,
	}

	err = r.collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("product not found")
		}
		return nil, err
	}

	return &product, nil
}

// FindAll lista productos con paginación y filtros
func (r *ProductRepository) FindAll(ctx context.Context, page, pageSize int, category, sortBy, sortOrder string, summary bool) ([]*models.Product, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Construir filtro
	filter := bson.M{"is_deleted": false}
	
	if category != "" {
		filter["category"] = category
	}

	// Contar total en paralelo
	totalCh := make(chan int64, 1)
	errCh := make(chan error, 1)
	
	go func() {
		total, err := r.collection.CountDocuments(ctx, filter)
		if err != nil {
			errCh <- err
			return
		}
		totalCh <- total
	}()

	// Opciones de búsqueda
	findOptions := options.Find()
	
	// Projection para listado resumido
	if summary {
		findOptions.SetProjection(bson.M{
			"sku":         1,
			"name":        1,
			"category":    1,
			"price_cents": 1,
			"currency":    1,
			"stock":       1,
			"images":      bson.M{"$slice": 1},
			"is_active":   1,
			"created_at":  1,
		})
	}
	
	// Paginación
	if page > 0 && pageSize > 0 {
		skip := (page - 1) * pageSize
		findOptions.SetSkip(int64(skip))
		findOptions.SetLimit(int64(pageSize))
	} else {
		findOptions.SetLimit(100)
	}
	
	// Ordenamiento
	sortField := "created_at"
	sortOrderInt := -1
	
	if sortBy != "" {
		sortField = sortBy
	}
	if sortOrder == "asc" {
		sortOrderInt = 1
	}
	
	findOptions.SetSort(bson.D{{Key: sortField, Value: sortOrderInt}})

	// Ejecutar query
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var products []*models.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}

	// Esperar el conteo
	var total int64
	select {
	case total = <-totalCh:
	case err := <-errCh:
		return products, 0, err
	case <-ctx.Done():
		return products, 0, ctx.Err()
	}

	return products, total, nil
}

// Update actualiza un producto
func (r *ProductRepository) Update(ctx context.Context, id string, update bson.M) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid product ID")
	}

	// Agregar updated_at automáticamente
	update["updated_at"] = time.Now()

	filter := bson.M{
		"_id":        objID,
		"is_deleted": false,
	}

	result, err := r.collection.UpdateOne(
		ctx,
		filter,
		bson.M{"$set": update},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// SoftDelete marca un producto como eliminado
func (r *ProductRepository) SoftDelete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid product ID")
	}

	filter := bson.M{
		"_id":        objID,
		"is_deleted": false,
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}