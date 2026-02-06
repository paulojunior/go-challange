package services

import (
	"context"
	"errors"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// mockProductRepository is a mock implementation of ProductRepository for testing.
type mockProductRepository struct {
	getAllProductsFunc   func(ctx context.Context, offset, limit int, filter models.ProductFilter) ([]models.Product, int64, error)
	getProductByCodeFunc func(ctx context.Context, code string) (*models.Product, error)
}

func (m *mockProductRepository) GetAllProducts(ctx context.Context, offset, limit int, filter models.ProductFilter) ([]models.Product, int64, error) {
	if m.getAllProductsFunc != nil {
		return m.getAllProductsFunc(ctx, offset, limit, filter)
	}
	return nil, 0, errors.New("not implemented")
}

func (m *mockProductRepository) GetProductByCode(ctx context.Context, code string) (*models.Product, error) {
	if m.getProductByCodeFunc != nil {
		return m.getProductByCodeFunc(ctx, code)
	}
	return nil, errors.New("not implemented")
}

func TestValidatePagination_Defaults(t *testing.T) {
	svc := NewCatalogService(&mockProductRepository{})

	params := svc.ValidatePagination(0, 0, false)

	if params.Offset != 0 {
		t.Errorf("expected default offset 0, got %d", params.Offset)
	}
	if params.Limit != 10 {
		t.Errorf("expected default limit 10, got %d", params.Limit)
	}
}

func TestValidatePagination_ValidValues(t *testing.T) {
	svc := NewCatalogService(&mockProductRepository{})

	params := svc.ValidatePagination(5, 20, true)

	if params.Offset != 5 {
		t.Errorf("expected offset 5, got %d", params.Offset)
	}
	if params.Limit != 20 {
		t.Errorf("expected limit 20, got %d", params.Limit)
	}
}

func TestValidatePagination_LimitValidation(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		limitProvided bool
		expectedLimit int
	}{
		{"limit not provided uses default", 0, false, 10},
		{"limit zero provided clamped to 1", 0, true, 1},
		{"limit below minimum clamped to 1", -5, true, 1},
		{"limit at minimum", 1, true, 1},
		{"limit above maximum clamped to 100", 200, true, 100},
		{"limit at maximum", 100, true, 100},
		{"valid limit", 50, true, 50},
	}

	svc := NewCatalogService(&mockProductRepository{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := svc.ValidatePagination(0, tt.limit, tt.limitProvided)

			if params.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, params.Limit)
			}
		})
	}
}

func TestValidatePagination_OffsetPassthrough(t *testing.T) {
	svc := NewCatalogService(&mockProductRepository{})

	// Service passes through offset as-is; negative offset validation
	// is handled at the handler layer (returns 400 Bad Request)
	params := svc.ValidatePagination(5, 10, true)

	if params.Offset != 5 {
		t.Errorf("expected offset 5, got %d", params.Offset)
	}
}

func TestListProducts_Success(t *testing.T) {
	mockRepo := &mockProductRepository{
		getAllProductsFunc: func(ctx context.Context, offset, limit int, filter models.ProductFilter) ([]models.Product, int64, error) {
			return []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(10.99),
					Category: &models.Category{
						Code: "CLOTHING",
						Name: "Clothing",
					},
				},
				{
					ID:    2,
					Code:  "PROD002",
					Price: decimal.NewFromFloat(20.50),
				},
			}, 2, nil
		},
	}

	svc := NewCatalogService(mockRepo)
	params := PaginationParams{Offset: 0, Limit: 10}
	filter := FilterParams{}

	result, err := svc.ListProducts(context.Background(), params, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}

	if len(result.Products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(result.Products))
	}

	// Verify first product with category
	if result.Products[0].Code != "PROD001" {
		t.Errorf("expected code PROD001, got %s", result.Products[0].Code)
	}
	if result.Products[0].Price != 10.99 {
		t.Errorf("expected price 10.99, got %f", result.Products[0].Price)
	}
	if result.Products[0].Category == nil {
		t.Fatal("expected category to be present")
	}
	if result.Products[0].Category.Code != "CLOTHING" {
		t.Errorf("expected category code CLOTHING, got %s", result.Products[0].Category.Code)
	}

	// Verify second product without category
	if result.Products[1].Category != nil {
		t.Error("expected category to be nil for second product")
	}
}

func TestListProducts_RepositoryError(t *testing.T) {
	mockRepo := &mockProductRepository{
		getAllProductsFunc: func(ctx context.Context, offset, limit int, filter models.ProductFilter) ([]models.Product, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	svc := NewCatalogService(mockRepo)
	params := PaginationParams{Offset: 0, Limit: 10}
	filter := FilterParams{}

	_, err := svc.ListProducts(context.Background(), params, filter)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetProductByCode_Success(t *testing.T) {
	variantPrice := decimal.NewFromFloat(11.99)
	mockRepo := &mockProductRepository{
		getProductByCodeFunc: func(ctx context.Context, code string) (*models.Product, error) {
			return &models.Product{
				ID:    1,
				Code:  "PROD001",
				Price: decimal.NewFromFloat(10.99),
				Category: &models.Category{
					Code: "CLOTHING",
					Name: "Clothing",
				},
				Variants: []models.Variant{
					{Name: "Small", SKU: "SKU001-S", Price: &variantPrice},
					{Name: "Large", SKU: "SKU001-L", Price: nil}, // nil = inherit product price
				},
			}, nil
		},
	}

	svc := NewCatalogService(mockRepo)

	result, err := svc.GetProductByCode(context.Background(), "PROD001")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Code != "PROD001" {
		t.Errorf("expected code PROD001, got %s", result.Code)
	}
	if result.Price != 10.99 {
		t.Errorf("expected price 10.99, got %f", result.Price)
	}
	if result.Category == nil {
		t.Fatal("expected category to be present")
	}
	if result.Category.Code != "CLOTHING" {
		t.Errorf("expected category code CLOTHING, got %s", result.Category.Code)
	}

	// Verify variants
	if len(result.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(result.Variants))
	}

	// First variant has its own price
	if result.Variants[0].Price != 11.99 {
		t.Errorf("expected variant price 11.99, got %f", result.Variants[0].Price)
	}

	// Second variant should inherit product price
	if result.Variants[1].Price != 10.99 {
		t.Errorf("expected variant to inherit product price 10.99, got %f", result.Variants[1].Price)
	}
}

func TestGetProductByCode_EmptyCode(t *testing.T) {
	mockRepo := &mockProductRepository{}

	svc := NewCatalogService(mockRepo)

	_, err := svc.GetProductByCode(context.Background(), "")

	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetProductByCode_NotFound(t *testing.T) {
	mockRepo := &mockProductRepository{
		getProductByCodeFunc: func(ctx context.Context, code string) (*models.Product, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	svc := NewCatalogService(mockRepo)

	_, err := svc.GetProductByCode(context.Background(), "INVALID")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetProductByCode_RepositoryError(t *testing.T) {
	mockRepo := &mockProductRepository{
		getProductByCodeFunc: func(ctx context.Context, code string) (*models.Product, error) {
			return nil, errors.New("database connection failed")
		},
	}

	svc := NewCatalogService(mockRepo)

	_, err := svc.GetProductByCode(context.Background(), "PROD001")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, ErrNotFound) {
		t.Error("expected generic error, not ErrNotFound")
	}
}

func TestGetProductByCode_NoCategory(t *testing.T) {
	mockRepo := &mockProductRepository{
		getProductByCodeFunc: func(ctx context.Context, code string) (*models.Product, error) {
			return &models.Product{
				ID:       1,
				Code:     "PROD001",
				Price:    decimal.NewFromFloat(10.99),
				Category: nil,
				Variants: []models.Variant{},
			}, nil
		},
	}

	svc := NewCatalogService(mockRepo)

	result, err := svc.GetProductByCode(context.Background(), "PROD001")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Category != nil {
		t.Error("expected category to be nil")
	}
}

func TestGetProductByCode_AllVariantsInheritPrice(t *testing.T) {
	mockRepo := &mockProductRepository{
		getProductByCodeFunc: func(ctx context.Context, code string) (*models.Product, error) {
			return &models.Product{
				ID:    1,
				Code:  "PROD001",
				Price: decimal.NewFromFloat(25.00),
				Variants: []models.Variant{
					{Name: "Red", SKU: "SKU-RED", Price: nil},   // nil = inherit product price
					{Name: "Blue", SKU: "SKU-BLUE", Price: nil}, // nil = inherit product price
				},
			}, nil
		},
	}

	svc := NewCatalogService(mockRepo)

	result, err := svc.GetProductByCode(context.Background(), "PROD001")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, v := range result.Variants {
		if v.Price != 25.00 {
			t.Errorf("variant %d: expected inherited price 25.00, got %f", i, v.Price)
		}
	}
}

func TestGetProductByCode_VariantWithZeroPrice(t *testing.T) {
	zeroPrice := decimal.NewFromFloat(0)
	mockRepo := &mockProductRepository{
		getProductByCodeFunc: func(ctx context.Context, code string) (*models.Product, error) {
			return &models.Product{
				ID:    1,
				Code:  "PROD001",
				Price: decimal.NewFromFloat(25.00),
				Variants: []models.Variant{
					{Name: "Free", SKU: "SKU-FREE", Price: &zeroPrice}, // Explicit 0.00 price
				},
			}, nil
		},
	}

	svc := NewCatalogService(mockRepo)

	result, err := svc.GetProductByCode(context.Background(), "PROD001")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Variant with explicit 0.00 price should NOT inherit product price
	if result.Variants[0].Price != 0.00 {
		t.Errorf("expected variant price 0.00, got %f", result.Variants[0].Price)
	}
}

func TestListProducts_WithCategoryFilter(t *testing.T) {
	mockRepo := &mockProductRepository{
		getAllProductsFunc: func(ctx context.Context, offset, limit int, filter models.ProductFilter) ([]models.Product, int64, error) {
			// Verify category filter is passed correctly
			if filter.Category != "CLOTHING" {
				t.Errorf("expected category filter CLOTHING, got %s", filter.Category)
			}
			return []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(10.99),
					Category: &models.Category{
						Code: "CLOTHING",
						Name: "Clothing",
					},
				},
			}, 1, nil
		},
	}

	svc := NewCatalogService(mockRepo)
	params := PaginationParams{Offset: 0, Limit: 10}
	filter := FilterParams{Category: "CLOTHING"}

	result, err := svc.ListProducts(context.Background(), params, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListProducts_WithPriceFilter(t *testing.T) {
	mockRepo := &mockProductRepository{
		getAllProductsFunc: func(ctx context.Context, offset, limit int, filter models.ProductFilter) ([]models.Product, int64, error) {
			// Verify price filter is passed correctly
			if filter.PriceLessThan == nil {
				t.Fatal("expected price filter to be set")
			}
			expected := decimal.NewFromInt(50)
			if !filter.PriceLessThan.Equal(expected) {
				t.Errorf("expected price filter 50, got %s", filter.PriceLessThan.String())
			}
			return []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(25.99),
				},
			}, 1, nil
		},
	}

	svc := NewCatalogService(mockRepo)
	params := PaginationParams{Offset: 0, Limit: 10}
	price := decimal.NewFromInt(50)
	filter := FilterParams{PriceLessThan: &price}

	result, err := svc.ListProducts(context.Background(), params, filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}
