package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product representa un producto en el catálogo
type Product struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SKU         string             `json:"sku" bson:"sku" binding:"required"`
	Name        string             `json:"name" bson:"name" binding:"required"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Category    string             `json:"category" bson:"category" binding:"required"`
	PriceCents  int                `json:"price_cents" bson:"price_cents" binding:"required"`
	Currency    string             `json:"currency" bson:"currency" binding:"required"`
	Stock       int                `json:"stock" bson:"stock"`
	Images      []string           `json:"images,omitempty" bson:"images,omitempty"`
	Attributes  map[string]string  `json:"attributes,omitempty" bson:"attributes,omitempty"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	IsDeleted   bool               `json:"-" bson:"is_deleted"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// ProductUpdate representa los campos actualizables de un producto
type ProductUpdate struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Category    *string            `json:"category,omitempty"`
	PriceCents  *int               `json:"price_cents,omitempty"`
	Currency    *string            `json:"currency,omitempty"`
	Stock       *int               `json:"stock,omitempty"`
	Images      []string           `json:"images,omitempty"`
	Attributes  map[string]string  `json:"attributes,omitempty"`
	IsActive    *bool              `json:"is_active,omitempty"`
}