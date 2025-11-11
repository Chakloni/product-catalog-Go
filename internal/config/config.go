package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
)

// GetMongoURI retorna la URI de MongoDB (configura aquí tu conexión)
func GetMongoURI() string {
	// Intentar obtener de variable de entorno primero
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	
	// REEMPLAZA ESTA LÍNEA CON TU CONEXIÓN DE MONGODB ATLAS
	return "mongodb+srv://<username>:<password>@cluster0.mongodb.net/?retryWrites=true&w=majority"
}

// InitDB inicializa la conexión a MongoDB con configuración optimizada
func InitDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	uri := GetMongoURI()

	// Opciones optimizadas del cliente
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).                   // Pool máximo de conexiones
		SetMinPoolSize(10).                    // Mantener conexiones mínimas activas
		SetMaxConnIdleTime(30 * time.Second).  // Limpiar conexiones inactivas
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetSocketTimeout(30 * time.Second).
		SetHeartbeatInterval(10 * time.Second).
		SetRetryWrites(true).                  // Reintentar escrituras fallidas
		SetRetryReads(true)                    // Reintentar lecturas fallidas

	// Conectar a MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	// Verificar la conexión
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error pinging MongoDB: %w", err)
	}

	Client = client
	// Database = client.Database("product_catalog")
	Database = os.Getenv("MONGO_DB")
	Collection = Database.Collection("products")

	log.Println("✅ Connected to MongoDB successfully")

	// Crear índices
	if err := createIndexes(ctx); err != nil {
		log.Printf("⚠️  Warning: Failed to create some indexes: %v", err)
	} else {
		log.Println("✅ Database indexes created successfully")
	}

	return nil
}

// createIndexes crea todos los índices necesarios para optimizar queries
func createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		// Índice único en SKU
		{
			Keys:    bson.D{{Key: "sku", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_sku_unique"),
		},
		// Índice en is_deleted (filtrado más común)
		{
			Keys:    bson.D{{Key: "is_deleted", Value: 1}},
			Options: options.Index().SetName("idx_is_deleted"),
		},
		// Índice compuesto para queries de listado
		{
			Keys: bson.D{
				{Key: "is_deleted", Value: 1},
				{Key: "is_active", Value: 1},
			},
			Options: options.Index().SetName("idx_deleted_active"),
		},
		// Índice en categoría para filtrado
		{
			Keys:    bson.D{{Key: "category", Value: 1}},
			Options: options.Index().SetName("idx_category"),
		},
		// Índice compuesto para filtrado por categoría
		{
			Keys: bson.D{
				{Key: "is_deleted", Value: 1},
				{Key: "category", Value: 1},
				{Key: "is_active", Value: 1},
			},
			Options: options.Index().SetName("idx_deleted_category_active"),
		},
		// Índice de texto para búsqueda
		{
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "description", Value: "text"},
			},
			Options: options.Index().
				SetName("idx_text_search").
				SetWeights(bson.M{"name": 10, "description": 5}),
		},
		// Índice en precio para ordenamiento
		{
			Keys:    bson.D{{Key: "price_cents", Value: 1}},
			Options: options.Index().SetName("idx_price"),
		},
		// Índice en stock para filtrar productos disponibles
		{
			Keys:    bson.D{{Key: "stock", Value: 1}},
			Options: options.Index().SetName("idx_stock"),
		},
		// Índice en created_at para ordenamiento temporal
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
	}

	// Crear índices
	_, err := Collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Close cierra la conexión a MongoDB de forma segura
func Close() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return Client.Disconnect(ctx)
	}
	return nil
}

// HealthCheck verifica que la conexión a la base de datos esté activa
func HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return Client.Ping(ctx, nil)
}