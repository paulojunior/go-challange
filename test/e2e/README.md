# End-to-End Tests

This directory contains end-to-end (e2e) tests for the application.

## Overview

E2E tests validate the entire application stack including:
- HTTP server and routing
- Database operations
- Request/response handling
- Integration between components

## Structure

```
test/e2e/
├── README.md           # This file
├── helpers.go          # Test utilities and setup helpers
├── catalog_test.go     # Catalog endpoints e2e tests
└── categories_test.go  # Categories endpoints e2e tests
```

## Running Tests

### Prerequisites

1. PostgreSQL database running
2. Environment variables configured (or use defaults):
   - `POSTGRES_USER` (default: postgres)
   - `POSTGRES_PASSWORD` (default: password)
   - `POSTGRES_DB_TEST` (default: go_challenge_test)
   - `POSTGRES_PORT` (default: 5432)

### Run all e2e tests

```bash
go test ./test/e2e/... -v
```

### Run specific test file

```bash
go test ./test/e2e/catalog_test.go ./test/e2e/helpers.go -v
go test ./test/e2e/categories_test.go ./test/e2e/helpers.go -v
```

### Run specific test

```bash
go test ./test/e2e/... -v -run TestCatalogEndpoint_ListProducts
go test ./test/e2e/... -v -run TestCategoriesEndpoint_CreateCategory
```

## Test Coverage

### Catalog Endpoints

**GET /v1/catalog**
- ✓ List all products with default pagination
- ✓ List products with custom pagination (offset, limit)
- ✓ Limit validation (min/max boundaries)
- ✓ Products include categories

**GET /v1/catalog/{code}**
- ✓ Get product by valid code
- ✓ Product details include category
- ✓ Product details include variants
- ✓ Variants inherit product price when not set
- ✓ Get product without variants
- ✓ Handle invalid product code (404)

**Integration**
- ✓ Full workflow: list products → get product details

### Categories Endpoints

**GET /v1/categories**
- ✓ List all categories
- ✓ List categories when empty
- ✓ Categories linked to products

**POST /v1/categories**
- ✓ Create new category
- ✓ Validation: missing code (400)
- ✓ Validation: missing name (400)
- ✓ Validation: empty values (400)

**Integration**
- ✓ Full workflow: create categories → list categories
- ✓ Categories exist for products

## Database Setup

Each test:
1. Creates a fresh test server instance
2. Connects to test database
3. Auto-migrates tables
4. Clears existing data
5. Seeds test data as needed
6. Runs test assertions
7. Cleans up resources

## Helpers

### TestServer
Main test server struct with:
- HTTP server instance
- Database connection
- Cleanup function

### Key Functions
- `SetupTestServer(t)` - Initialize test environment
- `ClearDatabase()` - Remove all test data
- `SeedCategories()` - Add test categories
- `SeedProducts()` - Add test products
- `GET(path)` - Make GET request
- `POST(path, body)` - Make POST request
- `DecodeJSON(resp, v)` - Parse JSON response
- `AssertStatusCode(t, expected, actual)` - Verify HTTP status
- `AssertNoError(t, err)` - Verify no errors

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Always defer cleanup to avoid resource leaks
3. **Seed Data**: Create only necessary test data
4. **Clear Assertions**: Use helper functions for consistent error messages
5. **Real Database**: Use actual database for realistic testing

## CI/CD Integration

To run e2e tests in CI:

```yaml
# Example GitHub Actions
- name: Run E2E Tests
  env:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
    POSTGRES_DB_TEST: test_db
    POSTGRES_PORT: 5432
  run: go test ./test/e2e/... -v
```

## Troubleshooting

### "connection refused"
- Ensure PostgreSQL is running
- Check connection parameters

### "database does not exist"
- Create test database: `createdb go_challenge_test`
- Or set `POSTGRES_DB_TEST` to existing database

### "permission denied"
- Verify database user has CREATE/DROP privileges
- Check POSTGRES_USER and POSTGRES_PASSWORD

## Future Enhancements

Potential additions:
- [ ] Filter tests (by category, price)
- [ ] Performance tests
- [ ] Concurrent request tests
- [ ] Error handling scenarios
- [ ] Authentication tests (when added)
