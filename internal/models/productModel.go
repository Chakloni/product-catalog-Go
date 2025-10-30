package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	SKU         string                 `json:"sku" bson:"sku"`
	Name        string                 `json:"name" bson:"name"`
	Description string                 `json:"description" bson:"description"`
	Category    string                 `json:"category" bson:"category"`
	PriceCents  int64                  `json:"price_cents" bson:"price_cents"`
	Currency    string                 `json:"currency" bson:"currency"`
	Stock       int64                  `json:"stock" bson:"stock"`
	Images      []string               `json:"images" bson:"images"`
	Attributes  map[string]interface{} `json:"attributes" bson:"attributes"`
	IsActive    bool                   `json:"is_active" bson:"is_active"`
	IsDeleted   bool                   `json:"is_deleted" bson:"is_deleted"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
}
