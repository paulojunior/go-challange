// Package services provides business logic layer for the application.
package services

import (
	"context"
	"errors"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PaginationParams holds validated pagination parameters.
type PaginationParams struct {
	Offset int
	Limit  int
}

// FilterParams holds filter criteria for product queries.
type FilterParams struct {
	Category      string
	PriceLessThan *decimal.Decimal
}

// ProductDTO represents a product for API responses.
type ProductDTO struct {
	Code     string
	Price    float64
	Category *CategoryDTO
}

// CategoryDTO represents a category for API responses.
type CategoryDTO struct {
	Code string
	Name string
}

// VariantDTO represents a variant for API responses.
type VariantDTO struct {
	Name  string
	SKU   string
	Price float64
}

// ProductDetailDTO represents detailed product information.
type ProductDetailDTO struct {
	Code     string
	Price    float64
	Category *CategoryDTO
	Variants []VariantDTO
}

// ProductListResult holds the result of listing products.
type ProductListResult struct {
	Products []ProductDTO
	Total    int64
}

// ProductRepository defines the interface for product data access.
type ProductRepository interface {
	GetAllProducts(ctx context.Context, offset, limit int, filter models.ProductFilter) ([]models.Product, int64, error)
	GetProductByCode(ctx context.Context, code string) (*models.Product, error)
}

// CatalogService handles catalog business logic.
type CatalogService struct {
	repo ProductRepository
}

// NewCatalogService creates a new CatalogService instance.
func NewCatalogService(repo ProductRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

// ValidatePagination validates and normalizes pagination parameters.
// Returns validated params with defaults: offset=0, limit=10.
// Limit is constrained between 1 and 100.
// Note: Negative offset validation is handled at the handler layer.
// The limitProvided flag indicates whether limit was explicitly set by the caller.
func (s *CatalogService) ValidatePagination(offset, limit int, limitProvided bool) PaginationParams {
	params := PaginationParams{
		Offset: offset,
		Limit:  10,
	}

	if limitProvided {
		// Limit was explicitly provided, clamp to valid range
		params.Limit = clamp(limit, 1, 100)
	}

	return params
}

// ListProducts retrieves paginated and filtered products.
func (s *CatalogService) ListProducts(ctx context.Context, params PaginationParams, filter FilterParams) (*ProductListResult, error) {
	repoFilter := models.ProductFilter{
		Category: filter.Category,
	}

	if filter.PriceLessThan != nil {
		repoFilter.PriceLessThan = filter.PriceLessThan
	}

	products, total, err := s.repo.GetAllProducts(ctx, params.Offset, params.Limit, repoFilter)
	if err != nil {
		return nil, err
	}

	result := &ProductListResult{
		Products: make([]ProductDTO, len(products)),
		Total:    total,
	}

	for i, p := range products {
		result.Products[i] = mapProductToDTO(p)
	}

	return result, nil
}

// GetProductByCode retrieves a product by its code.
// Returns ErrNotFound if the product doesn't exist.
func (s *CatalogService) GetProductByCode(ctx context.Context, code string) (*ProductDetailDTO, error) {
	if code == "" {
		return nil, ErrInvalidInput
	}

	product, err := s.repo.GetProductByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return mapProductToDetailDTO(product), nil
}

func mapProductToDTO(p models.Product) ProductDTO {
	dto := ProductDTO{
		Code:  p.Code,
		Price: p.Price.InexactFloat64(),
	}

	if p.Category != nil {
		dto.Category = &CategoryDTO{
			Code: p.Category.Code,
			Name: p.Category.Name,
		}
	}

	return dto
}

func mapProductToDetailDTO(p *models.Product) *ProductDetailDTO {
	detail := &ProductDetailDTO{
		Code:     p.Code,
		Price:    p.Price.InexactFloat64(),
		Variants: make([]VariantDTO, len(p.Variants)),
	}

	if p.Category != nil {
		detail.Category = &CategoryDTO{
			Code: p.Category.Code,
			Name: p.Category.Name,
		}
	}

	productPrice := p.Price.InexactFloat64()
	for i, v := range p.Variants {
		variantPrice := productPrice
		if v.Price != nil {
			variantPrice = v.Price.InexactFloat64()
		}

		detail.Variants[i] = VariantDTO{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: variantPrice,
		}
	}

	return detail
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
