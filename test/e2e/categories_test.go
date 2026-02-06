package e2e

import (
	"net/http"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/categories"
)

func TestCategoriesEndpoint_ListCategories(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Seed database
	AssertNoError(t, ts.ClearDatabase())
	AssertNoError(t, ts.SeedCategories())

	t.Run("list all categories", func(t *testing.T) {
		resp, err := ts.GET("/v1/categories")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response []categories.CategoryResponse
		AssertNoError(t, DecodeJSON(resp, &response))

		// Verify response
		if len(response) != 3 {
			t.Fatalf("expected 3 categories, got %d", len(response))
		}

		// Verify categories
		expectedCategories := map[string]string{
			"CLOTHING":    "Clothing",
			"SHOES":       "Shoes",
			"ACCESSORIES": "Accessories",
		}

		for _, cat := range response {
			if expectedName, ok := expectedCategories[cat.Code]; ok {
				if cat.Name != expectedName {
					t.Errorf("expected name %s for code %s, got %s", expectedName, cat.Code, cat.Name)
				}
			} else {
				t.Errorf("unexpected category code: %s", cat.Code)
			}
		}
	})

	t.Run("list categories when empty", func(t *testing.T) {
		AssertNoError(t, ts.ClearDatabase())

		resp, err := ts.GET("/v1/categories")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response []categories.CategoryResponse
		AssertNoError(t, DecodeJSON(resp, &response))

		// Should return empty array
		if len(response) != 0 {
			t.Errorf("expected 0 categories, got %d", len(response))
		}
	})
}

func TestCategoriesEndpoint_CreateCategory(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	AssertNoError(t, ts.ClearDatabase())

	t.Run("create new category", func(t *testing.T) {
		newCategory := categories.CreateCategoryRequest{
			Code: "ELECTRONICS",
			Name: "Electronics",
		}

		resp, err := ts.POST("/v1/categories", newCategory)
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusCreated, resp.StatusCode)

		var response categories.CategoryResponse
		AssertNoError(t, DecodeJSON(resp, &response))

		// Verify response
		if response.Code != "ELECTRONICS" {
			t.Errorf("expected code ELECTRONICS, got %s", response.Code)
		}

		if response.Name != "Electronics" {
			t.Errorf("expected name Electronics, got %s", response.Name)
		}

		// Verify it was actually created in database
		listResp, err := ts.GET("/v1/categories")
		AssertNoError(t, err)

		var categories []categories.CategoryResponse
		AssertNoError(t, DecodeJSON(listResp, &categories))

		if len(categories) != 1 {
			t.Errorf("expected 1 category in database, got %d", len(categories))
		}
	})

	t.Run("create category with missing code", func(t *testing.T) {
		invalidCategory := map[string]string{
			"name": "Invalid Category",
		}

		resp, err := ts.POST("/v1/categories", invalidCategory)
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create category with missing name", func(t *testing.T) {
		invalidCategory := map[string]string{
			"code": "INVALID",
		}

		resp, err := ts.POST("/v1/categories", invalidCategory)
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create category with empty values", func(t *testing.T) {
		invalidCategory := categories.CreateCategoryRequest{
			Code: "",
			Name: "",
		}

		resp, err := ts.POST("/v1/categories", invalidCategory)
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestCategoriesEndpoint_Integration(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	AssertNoError(t, ts.ClearDatabase())

	t.Run("full workflow - create and list categories", func(t *testing.T) {
		// 1. Verify empty list
		resp, err := ts.GET("/v1/categories")
		AssertNoError(t, err)

		var initialList []categories.CategoryResponse
		AssertNoError(t, DecodeJSON(resp, &initialList))

		if len(initialList) != 0 {
			t.Errorf("expected empty list, got %d categories", len(initialList))
		}

		// 2. Create first category
		firstCategory := categories.CreateCategoryRequest{
			Code: "CLOTHING",
			Name: "Clothing",
		}

		resp, err = ts.POST("/v1/categories", firstCategory)
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusCreated, resp.StatusCode)

		// 3. Create second category
		secondCategory := categories.CreateCategoryRequest{
			Code: "SHOES",
			Name: "Shoes",
		}

		resp, err = ts.POST("/v1/categories", secondCategory)
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusCreated, resp.StatusCode)

		// 4. List all categories
		resp, err = ts.GET("/v1/categories")
		AssertNoError(t, err)

		var finalList []categories.CategoryResponse
		AssertNoError(t, DecodeJSON(resp, &finalList))

		// Verify we have 2 categories
		if len(finalList) != 2 {
			t.Errorf("expected 2 categories, got %d", len(finalList))
		}

		// Verify both categories are present
		codes := make(map[string]bool)
		for _, cat := range finalList {
			codes[cat.Code] = true
		}

		if !codes["CLOTHING"] {
			t.Error("expected CLOTHING category to be present")
		}

		if !codes["SHOES"] {
			t.Error("expected SHOES category to be present")
		}
	})
}

func TestCategoriesEndpoint_WithProducts(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Seed database with categories and products
	AssertNoError(t, ts.ClearDatabase())
	AssertNoError(t, ts.SeedCategories())
	AssertNoError(t, ts.SeedProducts())

	t.Run("categories exist for products", func(t *testing.T) {
		// Get categories
		resp, err := ts.GET("/v1/categories")
		AssertNoError(t, err)

		var categoriesList []categories.CategoryResponse
		AssertNoError(t, DecodeJSON(resp, &categoriesList))

		// Get products
		resp, err = ts.GET("/v1/catalog")
		AssertNoError(t, err)

		var products struct {
			Products []struct {
				Code     string `json:"code"`
				Category *struct {
					Code string `json:"code"`
					Name string `json:"name"`
				} `json:"category"`
			} `json:"products"`
		}
		AssertNoError(t, DecodeJSON(resp, &products))

		// Verify all products have valid categories
		categoryMap := make(map[string]bool)
		for _, cat := range categoriesList {
			categoryMap[cat.Code] = true
		}

		for _, prod := range products.Products {
			if prod.Category != nil {
				if !categoryMap[prod.Category.Code] {
					t.Errorf("product %s has invalid category %s", prod.Code, prod.Category.Code)
				}
			}
		}
	})
}
