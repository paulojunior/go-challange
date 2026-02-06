package models

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductFilter holds filter criteria for product queries.
type ProductFilter struct {
	Category      string
	PriceLessThan *decimal.Decimal
}

// ProductsRepository provides database access for product operations.
type ProductsRepository struct {
	db *gorm.DB
}

// NewProductsRepository creates a new ProductsRepository instance.
func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

// GetAllProducts retrieves paginated products with their categories and variants.
// Results are ordered by ID for deterministic pagination.
func (r *ProductsRepository) GetAllProducts(ctx context.Context, offset, limit int, filter ProductFilter) ([]Product, int64, error) {
	var products []Product
	var total int64

	// Build base query with filters applied
	baseQuery := r.applyFilters(r.db.WithContext(ctx).Model(&Product{}), filter)

	// Get total count with filters applied
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated products with deterministic ordering
	findQuery := r.applyFilters(r.db.WithContext(ctx).Preload("Category").Preload("Variants"), filter)
	if err := findQuery.
		Order("products.id ASC").
		Offset(offset).
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// applyFilters applies filter criteria to a query.
// Note: Category filter uses exact match (case-sensitive) on category code.
func (r *ProductsRepository) applyFilters(query *gorm.DB, filter ProductFilter) *gorm.DB {
	if filter.Category != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", filter.Category)
	}

	if filter.PriceLessThan != nil {
		query = query.Where("products.price < ?", *filter.PriceLessThan)
	}

	return query
}

// GetProductByCode retrieves a product by its unique code.
func (r *ProductsRepository) GetProductByCode(ctx context.Context, code string) (*Product, error) {
	var product Product
	if err := r.db.WithContext(ctx).Preload("Category").Preload("Variants").
		Where("code = ?", code).
		First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
