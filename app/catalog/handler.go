// Package catalog provides HTTP handlers for product catalog endpoints.
package catalog

import (
	"context"
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/services"
	"github.com/shopspring/decimal"
)

// Response represents the paginated product list response.
type Response struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
}

// Category represents a category in API responses.
type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Product represents a product in API responses.
type Product struct {
	Code     string    `json:"code"`
	Price    float64   `json:"price"`
	Category *Category `json:"category,omitempty"`
}

// Variant represents a product variant in API responses.
type Variant struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

// ProductDetail represents detailed product information in API responses.
type ProductDetail struct {
	Code     string    `json:"code"`
	Price    float64   `json:"price"`
	Category *Category `json:"category,omitempty"`
	Variants []Variant `json:"variants"`
}

// CatalogService defines the interface for catalog business logic.
type CatalogService interface {
	ValidatePagination(offset, limit int, limitProvided bool) services.PaginationParams
	ListProducts(ctx context.Context, params services.PaginationParams, filter services.FilterParams) (*services.ProductListResult, error)
	GetProductByCode(ctx context.Context, code string) (*services.ProductDetailDTO, error)
}

// CatalogHandler handles HTTP requests for the catalog endpoints.
type CatalogHandler struct {
	service CatalogService
}

// NewCatalogHandler creates a new CatalogHandler instance.
func NewCatalogHandler(s CatalogService) *CatalogHandler {
	return &CatalogHandler{service: s}
}

// HandleGet handles GET /catalog requests for listing products.
// Supports query parameters: offset, limit, category, priceLessThan.
func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()

	// Parse and validate pagination
	offset, err := parseQueryIntWithValidation(query.Get("offset"))
	if err != nil {
		return services.ErrInvalidOffset
	}
	if offset < 0 {
		return services.ErrInvalidOffset
	}

	limit, limitProvided, err := parseQueryIntWithFlagAndValidation(query.Get("limit"))
	if err != nil {
		return services.ErrInvalidLimit
	}

	params := h.service.ValidatePagination(offset, limit, limitProvided)

	// Parse filters
	filter := services.FilterParams{
		Category: query.Get("category"),
	}

	if priceLessThanStr := query.Get("priceLessThan"); priceLessThanStr != "" {
		price, err := decimal.NewFromString(priceLessThanStr)
		if err != nil {
			return services.ErrInvalidPrice
		}
		if price.IsNegative() {
			return services.ErrNegativePrice
		}
		filter.PriceLessThan = &price
	}

	result, err := h.service.ListProducts(r.Context(), params, filter)
	if err != nil {
		return err
	}

	response := Response{
		Products: mapProductsToResponse(result.Products),
		Total:    result.Total,
	}

	api.OKResponse(w, r, response)
	return nil
}

// HandleGetByCode handles GET /catalog/{code} requests for product details.
func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) error {
	code := r.PathValue("code")

	detail, err := h.service.GetProductByCode(r.Context(), code)
	if err != nil {
		return err
	}

	response := mapDetailToResponse(detail)
	api.OKResponse(w, r, response)
	return nil
}

func mapProductsToResponse(products []services.ProductDTO) []Product {
	result := make([]Product, len(products))
	for i, p := range products {
		result[i] = Product{
			Code:  p.Code,
			Price: p.Price,
		}
		if p.Category != nil {
			result[i].Category = &Category{
				Code: p.Category.Code,
				Name: p.Category.Name,
			}
		}
	}
	return result
}

func mapDetailToResponse(detail *services.ProductDetailDTO) ProductDetail {
	response := ProductDetail{
		Code:     detail.Code,
		Price:    detail.Price,
		Variants: make([]Variant, len(detail.Variants)),
	}

	if detail.Category != nil {
		response.Category = &Category{
			Code: detail.Category.Code,
			Name: detail.Category.Name,
		}
	}

	for i, v := range detail.Variants {
		response.Variants[i] = Variant{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: v.Price,
		}
	}

	return response
}

// parseQueryIntWithValidation parses a query string parameter to int.
// Returns 0 for empty strings, or an error for invalid values.
func parseQueryIntWithValidation(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// parseQueryIntWithFlagAndValidation parses a query string parameter to int and indicates if it was provided.
// Returns (value, true, nil) if a valid integer was provided, (0, false, nil) if empty.
// Returns (0, false, error) if the value is invalid.
func parseQueryIntWithFlagAndValidation(s string) (int, bool, error) {
	if s == "" {
		return 0, false, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false, err
	}
	return v, true, nil
}
