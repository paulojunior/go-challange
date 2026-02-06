package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/services"
	"github.com/shopspring/decimal"
)

// mockCatalogService is a mock implementation of CatalogService for testing.
type mockCatalogService struct {
	validatePaginationFunc func(offset, limit int, limitProvided bool) services.PaginationParams
	listProductsFunc       func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error)
	getProductByCodeFunc   func(ctx context.Context, code string) (*services.ProductDetailDTO, error)
}

func (m *mockCatalogService) ValidatePagination(offset, limit int, limitProvided bool) services.PaginationParams {
	if m.validatePaginationFunc != nil {
		return m.validatePaginationFunc(offset, limit, limitProvided)
	}
	return services.PaginationParams{Offset: 0, Limit: 10}
}

func (m *mockCatalogService) ListProducts(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
	if m.listProductsFunc != nil {
		return m.listProductsFunc(ctx, params, filter)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCatalogService) GetProductByCode(ctx context.Context, code string) (*services.ProductDetailDTO, error) {
	if m.getProductByCodeFunc != nil {
		return m.getProductByCodeFunc(ctx, code)
	}
	return nil, errors.New("not implemented")
}

func TestHandleGetByCode_Success(t *testing.T) {
	// Setup mock service
	mockSvc := &mockCatalogService{
		getProductByCodeFunc: func(ctx context.Context, code string) (*services.ProductDetailDTO, error) {
			if code == "PROD001" {
				return &services.ProductDetailDTO{
					Code:  "PROD001",
					Price: 10.99,
					Category: &services.CategoryDTO{
						Code: "CLOTHING",
						Name: "Clothing",
					},
					Variants: []services.VariantDTO{
						{Name: "Variant A", SKU: "SKU001A", Price: 11.99},
						{Name: "Variant B", SKU: "SKU001B", Price: 10.99}, // Inherited price
					},
				}, nil
			}
			return nil, services.ErrNotFound
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
	req.SetPathValue("code", "PROD001")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGetByCode).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response ProductDetail
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify product details
	if response.Code != "PROD001" {
		t.Errorf("expected code PROD001, got %s", response.Code)
	}

	if response.Price != 10.99 {
		t.Errorf("expected price 10.99, got %f", response.Price)
	}

	// Verify category
	if response.Category == nil {
		t.Fatal("expected category to be present")
	}

	if response.Category.Code != "CLOTHING" {
		t.Errorf("expected category code CLOTHING, got %s", response.Category.Code)
	}

	if response.Category.Name != "Clothing" {
		t.Errorf("expected category name Clothing, got %s", response.Category.Name)
	}

	// Verify variants
	if len(response.Variants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(response.Variants))
	}

	// First variant should have its own price
	if response.Variants[0].Price != 11.99 {
		t.Errorf("expected variant A price 11.99, got %f", response.Variants[0].Price)
	}

	// Second variant should inherit product price
	if response.Variants[1].Price != 10.99 {
		t.Errorf("expected variant B to inherit product price 10.99, got %f", response.Variants[1].Price)
	}
}

func TestHandleGetByCode_ProductNotFound(t *testing.T) {
	// Setup mock service that returns not found error
	mockSvc := &mockCatalogService{
		getProductByCodeFunc: func(ctx context.Context, code string) (*services.ProductDetailDTO, error) {
			return nil, services.ErrNotFound
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/catalog/INVALID", nil)
	req.SetPathValue("code", "INVALID")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGetByCode).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleGetByCode_MissingCode(t *testing.T) {
	mockSvc := &mockCatalogService{
		getProductByCodeFunc: func(ctx context.Context, code string) (*services.ProductDetailDTO, error) {
			return nil, services.ErrInvalidInput
		},
	}
	handler := NewCatalogHandler(mockSvc)

	// Create request without code
	req := httptest.NewRequest(http.MethodGet, "/catalog/", nil)
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGetByCode).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleGetByCode_NoCategory(t *testing.T) {
	// Setup mock service with product without category
	mockSvc := &mockCatalogService{
		getProductByCodeFunc: func(ctx context.Context, code string) (*services.ProductDetailDTO, error) {
			return &services.ProductDetailDTO{
				Code:     "PROD001",
				Price:    10.99,
				Category: nil, // No category
				Variants: []services.VariantDTO{},
			}, nil
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
	req.SetPathValue("code", "PROD001")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGetByCode).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response ProductDetail
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify category is nil
	if response.Category != nil {
		t.Error("expected category to be nil")
	}
}

func TestHandleGet_WithPagination(t *testing.T) {
	// Setup mock service
	mockSvc := &mockCatalogService{
		validatePaginationFunc: func(offset, limit int, limitProvided bool) services.PaginationParams {
			return services.PaginationParams{Offset: 5, Limit: 20}
		},
		listProductsFunc: func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
			// Verify pagination parameters are passed correctly
			if params.Offset != 5 || params.Limit != 20 {
				t.Errorf("expected offset=5, limit=20, got offset=%d, limit=%d", params.Offset, params.Limit)
			}

			return &services.ProductListResult{
				Products: []services.ProductDTO{
					{
						Code:  "PROD006",
						Price: 5.50,
						Category: &services.CategoryDTO{
							Code: "SHOES",
							Name: "Shoes",
						},
					},
				},
				Total: 8, // Total of 8 products
			}, nil
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request with pagination parameters
	req := httptest.NewRequest(http.MethodGet, "/catalog?offset=5&limit=20", nil)
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify total
	if response.Total != 8 {
		t.Errorf("expected total 8, got %d", response.Total)
	}

	// Verify products
	if len(response.Products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(response.Products))
	}
}

func TestHandleGet_DefaultPagination(t *testing.T) {
	// Setup mock service
	mockSvc := &mockCatalogService{
		validatePaginationFunc: func(offset, limit int, limitProvided bool) services.PaginationParams {
			// Return default values
			return services.PaginationParams{Offset: 0, Limit: 10}
		},
		listProductsFunc: func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
			// Verify default values are used
			if params.Offset != 0 {
				t.Errorf("expected default offset=0, got %d", params.Offset)
			}
			if params.Limit != 10 {
				t.Errorf("expected default limit=10, got %d", params.Limit)
			}

			return &services.ProductListResult{
				Products: []services.ProductDTO{},
				Total:    0,
			}, nil
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request without pagination parameters
	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleGet_WithCategory(t *testing.T) {
	// Setup mock service
	mockSvc := &mockCatalogService{
		validatePaginationFunc: func(offset, limit int, limitProvided bool) services.PaginationParams {
			return services.PaginationParams{Offset: 0, Limit: 10}
		},
		listProductsFunc: func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
			return &services.ProductListResult{
				Products: []services.ProductDTO{
					{
						Code:  "PROD001",
						Price: 10.99,
						Category: &services.CategoryDTO{
							Code: "CLOTHING",
							Name: "Clothing",
						},
					},
				},
				Total: 1,
			}, nil
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify category is included
	if response.Products[0].Category == nil {
		t.Fatal("expected category to be present")
	}

	if response.Products[0].Category.Code != "CLOTHING" {
		t.Errorf("expected category code CLOTHING, got %s", response.Products[0].Category.Code)
	}
}

func TestHandleGet_RepositoryError(t *testing.T) {
	// Setup mock service that returns error
	mockSvc := &mockCatalogService{
		validatePaginationFunc: func(offset, limit int, limitProvided bool) services.PaginationParams {
			return services.PaginationParams{Offset: 0, Limit: 10}
		},
		listProductsFunc: func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHandleGetByCode_InternalError(t *testing.T) {
	// Setup mock service that returns internal error (not ErrNotFound)
	mockSvc := &mockCatalogService{
		getProductByCodeFunc: func(ctx context.Context, code string) (*services.ProductDetailDTO, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewCatalogHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
	req.SetPathValue("code", "PROD001")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGetByCode).ServeHTTP(w, req)

	// Assert response - should be 500, not 404
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestParseQueryIntWithValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{"empty string returns 0", "", 0, false},
		{"valid positive number", "42", 42, false},
		{"zero", "0", 0, false},
		{"negative number", "-5", -5, false},
		{"invalid string returns error", "abc", 0, true},
		{"mixed string returns error", "12abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQueryIntWithValidation(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("parseQueryIntWithValidation(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseQueryIntWithValidation(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseQueryIntWithValidation(%q) = %d, expected %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestParseQueryIntWithFlagAndValidation(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedValue    int
		expectedProvided bool
		expectError      bool
	}{
		{"empty string returns not provided", "", 0, false, false},
		{"valid positive number", "42", 42, true, false},
		{"zero is provided", "0", 0, true, false},
		{"negative number", "-5", -5, true, false},
		{"invalid string returns error", "abc", 0, false, true},
		{"mixed string returns error", "12abc", 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, provided, err := parseQueryIntWithFlagAndValidation(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("parseQueryIntWithFlagAndValidation(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseQueryIntWithFlagAndValidation(%q) unexpected error: %v", tt.input, err)
			}
			if value != tt.expectedValue {
				t.Errorf("parseQueryIntWithFlagAndValidation(%q) value = %d, expected %d", tt.input, value, tt.expectedValue)
			}
			if provided != tt.expectedProvided {
				t.Errorf("parseQueryIntWithFlagAndValidation(%q) provided = %v, expected %v", tt.input, provided, tt.expectedProvided)
			}
		})
	}
}

func TestHandleGet_WithCategoryFilter(t *testing.T) {
	mockSvc := &mockCatalogService{
		validatePaginationFunc: func(offset, limit int, limitProvided bool) services.PaginationParams {
			return services.PaginationParams{Offset: 0, Limit: 10}
		},
		listProductsFunc: func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
			// Verify category filter is passed correctly
			if filter.Category != "CLOTHING" {
				t.Errorf("expected category filter CLOTHING, got %s", filter.Category)
			}
			return &services.ProductListResult{
				Products: []services.ProductDTO{},
				Total:    0,
			}, nil
		},
	}

	handler := NewCatalogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/catalog?category=CLOTHING", nil)
	w := httptest.NewRecorder()

	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleGet_WithPriceFilter(t *testing.T) {
	mockSvc := &mockCatalogService{
		validatePaginationFunc: func(offset, limit int, limitProvided bool) services.PaginationParams {
			return services.PaginationParams{Offset: 0, Limit: 10}
		},
		listProductsFunc: func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
			// Verify price filter is passed correctly
			if filter.PriceLessThan == nil {
				t.Fatal("expected price filter to be set")
			}
			expected := decimal.NewFromInt(50)
			if !filter.PriceLessThan.Equal(expected) {
				t.Errorf("expected price filter 50, got %s", filter.PriceLessThan.String())
			}
			return &services.ProductListResult{
				Products: []services.ProductDTO{},
				Total:    0,
			}, nil
		},
	}

	handler := NewCatalogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/catalog?priceLessThan=50", nil)
	w := httptest.NewRecorder()

	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleGet_LimitProvidedFlag(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		expectedProvided bool
	}{
		{"limit not provided", "/catalog", false},
		{"limit provided as 0", "/catalog?limit=0", true},
		{"limit provided as 10", "/catalog?limit=10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockCatalogService{
				validatePaginationFunc: func(offset, limit int, limitProvided bool) services.PaginationParams {
					if limitProvided != tt.expectedProvided {
						t.Errorf("expected limitProvided=%v, got %v", tt.expectedProvided, limitProvided)
					}
					return services.PaginationParams{Offset: 0, Limit: 10}
				},
				listProductsFunc: func(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error) {
					return &services.ProductListResult{Products: []services.ProductDTO{}, Total: 0}, nil
				},
			}

			handler := NewCatalogHandler(mockSvc)
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}

func TestHandleGet_InvalidPriceFilter(t *testing.T) {
	mockSvc := &mockCatalogService{}

	handler := NewCatalogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/catalog?priceLessThan=abc", nil)
	w := httptest.NewRecorder()

	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleGet_NegativePriceFilter(t *testing.T) {
	mockSvc := &mockCatalogService{}

	handler := NewCatalogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/catalog?priceLessThan=-10", nil)
	w := httptest.NewRecorder()

	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleGet_InvalidOffset(t *testing.T) {
	mockSvc := &mockCatalogService{}

	handler := NewCatalogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/catalog?offset=abc", nil)
	w := httptest.NewRecorder()

	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleGet_InvalidLimit(t *testing.T) {
	mockSvc := &mockCatalogService{}

	handler := NewCatalogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/catalog?limit=abc", nil)
	w := httptest.NewRecorder()

	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleGet_NegativeOffset(t *testing.T) {
	mockSvc := &mockCatalogService{}

	handler := NewCatalogHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/catalog?offset=-5", nil)
	w := httptest.NewRecorder()

	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
