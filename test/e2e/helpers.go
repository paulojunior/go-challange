// Package e2e provides end-to-end testing utilities and helpers.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/app/services"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TestServer represents a test HTTP server with database.
type TestServer struct {
	Server    *httptest.Server
	DB        *gorm.DB
	CleanupFn func()
}

// SetupTestServer creates a test server with a PostgreSQL test database.
func SetupTestServer(t *testing.T) *TestServer {
	// Use test database configuration.
	db, cleanup, err := database.New(
		getEnv("POSTGRES_USER", "postgres"),
		getEnv("POSTGRES_PASSWORD", "password"),
		getEnv("POSTGRES_DB_TEST", "go_challenge_test"),
		getEnv("POSTGRES_PORT", "5432"),
	)
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate tables.
	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.Variant{}); err != nil {
		t.Fatalf("failed to auto-migrate tables: %v", err)
	}

	// Initialize repositories.
	prodRepo := models.NewProductsRepository(db)
	catRepo := models.NewCategoriesRepository(db)

	// Initialize services.
	catalogService := services.NewCatalogService(prodRepo)
	categoriesService := services.NewCategoriesService(catRepo)

	// Initialize handlers.
	catHandler := catalog.NewCatalogHandler(catalogService)
	categoriesHandler := categories.NewCategoriesHandler(categoriesService)

	// Set up routing.
	mux := http.NewServeMux()
	mux.Handle("GET /catalog", api.ErrorHandler(catHandler.HandleGet))
	mux.Handle("GET /catalog/{code}", api.ErrorHandler(catHandler.HandleGetByCode))
	mux.Handle("GET /categories", api.ErrorHandler(categoriesHandler.HandleGet))
	mux.Handle("POST /categories", api.ErrorHandler(categoriesHandler.HandlePost))

	// Create test server.
	server := httptest.NewServer(mux)

	return &TestServer{
		Server: server,
		DB:     db,
		CleanupFn: func() {
			server.Close()
			cleanup()
		},
	}
}

// Cleanup closes the test server and database.
func (ts *TestServer) Cleanup() {
	if ts.CleanupFn != nil {
		ts.CleanupFn()
	}
}

// ClearDatabase clears all data from test database.
func (ts *TestServer) ClearDatabase() error {
	// Delete in order to respect foreign keys.
	if err := ts.DB.Exec("DELETE FROM product_variants").Error; err != nil {
		return err
	}
	if err := ts.DB.Exec("DELETE FROM products").Error; err != nil {
		return err
	}
	if err := ts.DB.Exec("DELETE FROM categories").Error; err != nil {
		return err
	}
	return nil
}

// SeedCategories adds test categories to the database.
func (ts *TestServer) SeedCategories() error {
	categories := []models.Category{
		{Code: "CLOTHING", Name: "Clothing"},
		{Code: "SHOES", Name: "Shoes"},
		{Code: "ACCESSORIES", Name: "Accessories"},
	}

	for _, cat := range categories {
		if err := ts.DB.Create(&cat).Error; err != nil {
			return err
		}
	}
	return nil
}

// SeedProducts adds test products to the database.
func (ts *TestServer) SeedProducts() error {
	// Get category IDs.
	var clothing, shoes, accessories models.Category
	if err := ts.DB.Where("code = ?", "CLOTHING").First(&clothing).Error; err != nil {
		return err
	}
	if err := ts.DB.Where("code = ?", "SHOES").First(&shoes).Error; err != nil {
		return err
	}
	if err := ts.DB.Where("code = ?", "ACCESSORIES").First(&accessories).Error; err != nil {
		return err
	}

	variantAPrice := decimal.NewFromFloat(11.99)
	products := []models.Product{
		{
			Code:       "PROD001",
			Price:      decimal.NewFromFloat(10.99),
			CategoryID: &clothing.ID,
			Variants: []models.Variant{
				{Name: "Variant A", SKU: "SKU001A", Price: &variantAPrice},
				{Name: "Variant B", SKU: "SKU001B", Price: nil}, // nil = inherit product price
			},
		},
		{
			Code:       "PROD002",
			Price:      decimal.NewFromFloat(12.49),
			CategoryID: &shoes.ID,
			Variants:   []models.Variant{},
		},
		{
			Code:       "PROD003",
			Price:      decimal.NewFromFloat(8.75),
			CategoryID: &accessories.ID,
			Variants:   []models.Variant{},
		},
	}

	for _, prod := range products {
		if err := ts.DB.Create(&prod).Error; err != nil {
			return err
		}
	}
	return nil
}

// GET makes a GET request to the test server.
func (ts *TestServer) GET(path string) (*http.Response, error) {
	return http.Get(ts.Server.URL + path)
}

// POST makes a POST request to the test server.
func (ts *TestServer) POST(path string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return http.Post(
		ts.Server.URL+path,
		"application/json",
		bytes.NewReader(jsonBody),
	)
}

// DecodeJSON decodes JSON response body.
func DecodeJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// getEnv gets environment variable with fallback.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// AssertStatusCode asserts the HTTP status code.
func AssertStatusCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected status code %d, got %d", expected, actual)
	}
}

// AssertNoError asserts no error occurred.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// PrintResponse prints the response body for debugging.
func PrintResponse(resp *http.Response) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read response body: %v\n", err)
		return
	}
	fmt.Printf("Response: %s\n", string(body))
}
