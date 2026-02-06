// Package categories provides HTTP handlers for category management endpoints.
package categories

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/services"
)

// CategoryResponse represents a category in API responses.
type CategoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CreateCategoryRequest represents the request body for creating a category.
type CreateCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CategoriesService defines the interface for category business logic.
type CategoriesService interface {
	ListCategories(ctx context.Context) ([]services.CategoryDTO, error)
	CreateCategory(ctx context.Context, input services.CreateCategoryInput) (*services.CategoryDTO, error)
}

// CategoriesHandler handles HTTP requests for the categories endpoints.
type CategoriesHandler struct {
	service CategoriesService
}

// NewCategoriesHandler creates a new CategoriesHandler instance.
func NewCategoriesHandler(s CategoriesService) *CategoriesHandler {
	return &CategoriesHandler{service: s}
}

// HandleGet handles GET /categories requests for listing categories.
func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) error {
	categories, err := h.service.ListCategories(r.Context())
	if err != nil {
		return err
	}

	response := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		response[i] = CategoryResponse{
			Code: c.Code,
			Name: c.Name,
		}
	}

	api.OKResponse(w, r, response)
	return nil
}

// HandlePost handles POST /categories requests for creating a category.
func (h *CategoriesHandler) HandlePost(w http.ResponseWriter, r *http.Request) error {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return services.ErrInvalidInput
	}

	input := services.CreateCategoryInput{
		Code: req.Code,
		Name: req.Name,
	}

	category, err := h.service.CreateCategory(r.Context(), input)
	if err != nil {
		return err
	}

	response := CategoryResponse{
		Code: category.Code,
		Name: category.Name,
	}

	api.CreatedResponse(w, r, response)
	return nil
}
