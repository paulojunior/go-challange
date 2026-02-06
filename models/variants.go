package models

import (
	"github.com/shopspring/decimal"
)

// Variant represents a product variant in the catalog.
// It includes a unique name, SKU, and an optional price.
// When Price is nil, the variant inherits the product's base price.
// When Price is set (even to 0.00), that value is used as the variant's price.
type Variant struct {
	ID        uint             `gorm:"primaryKey"`
	ProductID uint             `gorm:"not null"`
	Name      string           `gorm:"not null"`
	SKU       string           `gorm:"uniqueIndex;not null"`
	Price     *decimal.Decimal `gorm:"type:decimal(10,2);null"`
}

// TableName returns the database table name for Variant.
func (v *Variant) TableName() string {
	return "product_variants"
}
