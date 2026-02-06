package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/services"
)

// mockCategoriesService is a mock implementation of CategoriesService for testing.
type mockCategoriesService struct {
	listCategoriesFunc func(ctx context.Context) ([]services.CategoryDTO, error)
	createCategoryFunc func(ctx context.Context, input services.CreateCategoryInput) (*services.CategoryDTO, error)
}

func (m *mockCategoriesService) ListCategories(ctx context.Context) ([]services.CategoryDTO, error) {
	if m.listCategoriesFunc != nil {
		return m.listCategoriesFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCategoriesService) CreateCategory(ctx context.Context, input services.CreateCategoryInput) (*services.CategoryDTO, error) {
	if m.createCategoryFunc != nil {
		return m.createCategoryFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func TestHandleGet_Success(t *testing.T) {
	// Setup mock service
	mockSvc := &mockCategoriesService{
		listCategoriesFunc: func(ctx context.Context) ([]services.CategoryDTO, error) {
			return []services.CategoryDTO{
				{Code: "CLOTHING", Name: "Clothing"},
				{Code: "SHOES", Name: "Shoes"},
				{Code: "ACCESSORIES", Name: "Accessories"},
			}, nil
		},
	}

	handler := NewCategoriesHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []CategoryResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify response
	if len(response) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(response))
	}

	if response[0].Code != "CLOTHING" {
		t.Errorf("expected first category code CLOTHING, got %s", response[0].Code)
	}

	if response[1].Name != "Shoes" {
		t.Errorf("expected second category name Shoes, got %s", response[1].Name)
	}
}

func TestHandleGet_RepositoryError(t *testing.T) {
	// Setup mock service that returns error
	mockSvc := &mockCategoriesService{
		listCategoriesFunc: func(ctx context.Context) ([]services.CategoryDTO, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewCategoriesHandler(mockSvc)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandleGet).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHandlePost_Success(t *testing.T) {
	// Setup mock service
	mockSvc := &mockCategoriesService{
		createCategoryFunc: func(ctx context.Context, input services.CreateCategoryInput) (*services.CategoryDTO, error) {
			return &services.CategoryDTO{
				Code: input.Code,
				Name: input.Name,
			}, nil
		},
	}

	handler := NewCategoriesHandler(mockSvc)

	// Create request
	reqBody := CreateCategoryRequest{
		Code: "ELECTRONICS",
		Name: "Electronics",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandlePost).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response CategoryResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify response
	if response.Code != "ELECTRONICS" {
		t.Errorf("expected code ELECTRONICS, got %s", response.Code)
	}

	if response.Name != "Electronics" {
		t.Errorf("expected name Electronics, got %s", response.Name)
	}
}

func TestHandlePost_InvalidJSON(t *testing.T) {
	mockSvc := &mockCategoriesService{}
	handler := NewCategoriesHandler(mockSvc)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandlePost).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePost_MissingCode(t *testing.T) {
	mockSvc := &mockCategoriesService{
		createCategoryFunc: func(ctx context.Context, input services.CreateCategoryInput) (*services.CategoryDTO, error) {
			return nil, services.ErrInvalidInput
		},
	}
	handler := NewCategoriesHandler(mockSvc)

	// Create request with missing code
	reqBody := CreateCategoryRequest{
		Name: "Electronics",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandlePost).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePost_MissingName(t *testing.T) {
	mockSvc := &mockCategoriesService{
		createCategoryFunc: func(ctx context.Context, input services.CreateCategoryInput) (*services.CategoryDTO, error) {
			return nil, services.ErrInvalidInput
		},
	}
	handler := NewCategoriesHandler(mockSvc)

	// Create request with missing name
	reqBody := CreateCategoryRequest{
		Code: "ELECTRONICS",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandlePost).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePost_RepositoryError(t *testing.T) {
	// Setup mock service that returns error
	mockSvc := &mockCategoriesService{
		createCategoryFunc: func(ctx context.Context, input services.CreateCategoryInput) (*services.CategoryDTO, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewCategoriesHandler(mockSvc)

	// Create request
	reqBody := CreateCategoryRequest{
		Code: "ELECTRONICS",
		Name: "Electronics",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	api.ErrorHandler(handler.HandlePost).ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
