package models

import (
	"context"

	"gorm.io/gorm"
)

// CategoriesRepository provides database access for category operations.
type CategoriesRepository struct {
	db *gorm.DB
}

// NewCategoriesRepository creates a new CategoriesRepository instance.
func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{
		db: db,
	}
}

// GetAllCategories retrieves all categories from the database.
func (r *CategoriesRepository) GetAllCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateCategory creates a new category with the given code and name.
func (r *CategoriesRepository) CreateCategory(ctx context.Context, code, name string) (*Category, error) {
	category := Category{
		Code: code,
		Name: name,
	}

	if err := r.db.WithContext(ctx).Create(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}
