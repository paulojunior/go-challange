package e2e

import (
	"net/http"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/catalog"
)

func TestCatalogEndpoint_ListProducts(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Seed database
	AssertNoError(t, ts.ClearDatabase())
	AssertNoError(t, ts.SeedCategories())
	AssertNoError(t, ts.SeedProducts())

	t.Run("list all products with default pagination", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// Verify response
		if response.Total != 3 {
			t.Errorf("expected total 3, got %d", response.Total)
		}

		if len(response.Products) != 3 {
			t.Errorf("expected 3 products, got %d", len(response.Products))
		}

		// Verify first product has category
		if response.Products[0].Category == nil {
			t.Error("expected product to have category")
		}
	})

	t.Run("list products with pagination", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?offset=1&limit=2")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// Verify response
		if response.Total != 3 {
			t.Errorf("expected total 3, got %d", response.Total)
		}

		if len(response.Products) != 2 {
			t.Errorf("expected 2 products, got %d", len(response.Products))
		}
	})

	t.Run("list products with limit validation", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?limit=200")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// Should still return all products even with high limit
		if len(response.Products) != 3 {
			t.Errorf("expected 3 products, got %d", len(response.Products))
		}
	})
}

func TestCatalogEndpoint_GetProductByCode(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Seed database
	AssertNoError(t, ts.ClearDatabase())
	AssertNoError(t, ts.SeedCategories())
	AssertNoError(t, ts.SeedProducts())

	t.Run("get product by valid code", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog/PROD001")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.ProductDetail
		AssertNoError(t, DecodeJSON(resp, &response))

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
			t.Errorf("expected category CLOTHING, got %s", response.Category.Code)
		}

		// Verify variants
		if len(response.Variants) != 2 {
			t.Fatalf("expected 2 variants, got %d", len(response.Variants))
		}

		// First variant should have its own price
		if response.Variants[0].Price != 11.99 {
			t.Errorf("expected variant price 11.99, got %f", response.Variants[0].Price)
		}

		// Second variant should inherit product price (0 becomes 10.99)
		if response.Variants[1].Price != 10.99 {
			t.Errorf("expected variant to inherit price 10.99, got %f", response.Variants[1].Price)
		}
	})

	t.Run("get product without variants", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog/PROD002")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.ProductDetail
		AssertNoError(t, DecodeJSON(resp, &response))

		// Verify product
		if response.Code != "PROD002" {
			t.Errorf("expected code PROD002, got %s", response.Code)
		}

		// Verify no variants
		if len(response.Variants) != 0 {
			t.Errorf("expected 0 variants, got %d", len(response.Variants))
		}

		// Verify category
		if response.Category == nil {
			t.Fatal("expected category to be present")
		}

		if response.Category.Code != "SHOES" {
			t.Errorf("expected category SHOES, got %s", response.Category.Code)
		}
	})

	t.Run("get product with invalid code", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog/INVALID")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestCatalogEndpoint_Integration(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Seed database
	AssertNoError(t, ts.ClearDatabase())
	AssertNoError(t, ts.SeedCategories())
	AssertNoError(t, ts.SeedProducts())

	t.Run("full workflow - list and get details", func(t *testing.T) {
		// 1. List products
		resp, err := ts.GET("/v1/catalog")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var listResponse catalog.Response
		AssertNoError(t, DecodeJSON(resp, &listResponse))

		// Verify we got products
		if len(listResponse.Products) == 0 {
			t.Fatal("expected at least one product")
		}

		// 2. Get details of first product
		firstProductCode := listResponse.Products[0].Code
		resp, err = ts.GET("/v1/catalog/" + firstProductCode)
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var detailResponse catalog.ProductDetail
		AssertNoError(t, DecodeJSON(resp, &detailResponse))

		// Verify details match
		if detailResponse.Code != firstProductCode {
			t.Errorf("expected code %s, got %s", firstProductCode, detailResponse.Code)
		}

		if detailResponse.Price != listResponse.Products[0].Price {
			t.Errorf("expected price %f, got %f", listResponse.Products[0].Price, detailResponse.Price)
		}
	})
}

func TestCatalogEndpoint_Filters(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Seed database
	AssertNoError(t, ts.ClearDatabase())
	AssertNoError(t, ts.SeedCategories())
	AssertNoError(t, ts.SeedProducts())

	t.Run("filter by category", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?category=CLOTHING")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// Only PROD001 is in CLOTHING category
		if response.Total != 1 {
			t.Errorf("expected total 1, got %d", response.Total)
		}

		if len(response.Products) != 1 {
			t.Fatalf("expected 1 product, got %d", len(response.Products))
		}

		if response.Products[0].Code != "PROD001" {
			t.Errorf("expected PROD001, got %s", response.Products[0].Code)
		}

		if response.Products[0].Category.Code != "CLOTHING" {
			t.Errorf("expected category CLOTHING, got %s", response.Products[0].Category.Code)
		}
	})

	t.Run("filter by priceLessThan", func(t *testing.T) {
		// Products: PROD001=10.99, PROD002=12.49, PROD003=8.75
		resp, err := ts.GET("/v1/catalog?priceLessThan=11")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// PROD001 (10.99) and PROD003 (8.75) are less than 11
		if response.Total != 2 {
			t.Errorf("expected total 2, got %d", response.Total)
		}

		if len(response.Products) != 2 {
			t.Errorf("expected 2 products, got %d", len(response.Products))
		}

		// Verify all products have price < 11
		for _, p := range response.Products {
			if p.Price >= 11 {
				t.Errorf("product %s has price %f, expected less than 11", p.Code, p.Price)
			}
		}
	})

	t.Run("filter by category and priceLessThan combined", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?category=CLOTHING&priceLessThan=15")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// Only PROD001 matches (CLOTHING and price 10.99 < 15)
		if response.Total != 1 {
			t.Errorf("expected total 1, got %d", response.Total)
		}

		if len(response.Products) != 1 {
			t.Fatalf("expected 1 product, got %d", len(response.Products))
		}

		if response.Products[0].Code != "PROD001" {
			t.Errorf("expected PROD001, got %s", response.Products[0].Code)
		}
	})

	t.Run("filter with no matches", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?category=NONEXISTENT")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		if response.Total != 0 {
			t.Errorf("expected total 0, got %d", response.Total)
		}

		if len(response.Products) != 0 {
			t.Errorf("expected 0 products, got %d", len(response.Products))
		}
	})

	t.Run("invalid priceLessThan returns bad request", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?priceLessThan=abc")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestCatalogEndpoint_LimitZero(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Seed database
	AssertNoError(t, ts.ClearDatabase())
	AssertNoError(t, ts.SeedCategories())
	AssertNoError(t, ts.SeedProducts())

	t.Run("limit=0 clamps to 1", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?limit=0")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// Total should still be 3 (all products)
		if response.Total != 3 {
			t.Errorf("expected total 3, got %d", response.Total)
		}

		// But only 1 product should be returned (limit clamped to 1)
		if len(response.Products) != 1 {
			t.Errorf("expected 1 product (limit=0 clamped to 1), got %d", len(response.Products))
		}
	})

	t.Run("limit not provided uses default 10", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusOK, resp.StatusCode)

		var response catalog.Response
		AssertNoError(t, DecodeJSON(resp, &response))

		// All 3 products should be returned (default limit 10)
		if len(response.Products) != 3 {
			t.Errorf("expected 3 products, got %d", len(response.Products))
		}
	})
}

func TestCatalogEndpoint_InvalidPagination(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Run("invalid offset returns bad request", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?offset=abc")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid limit returns bad request", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?limit=abc")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("negative offset returns bad request", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?offset=-5")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("negative priceLessThan returns bad request", func(t *testing.T) {
		resp, err := ts.GET("/v1/catalog?priceLessThan=-10")
		AssertNoError(t, err)
		AssertStatusCode(t, http.StatusBadRequest, resp.StatusCode)
	})
}
