# Go Hiring Challenge

This repository contains a Go application for managing products and their prices, including functionalities for CRUD operations and seeding the database with initial data.

## Project Structure

```
.
├── cmd/
│   ├── server/main.go      # HTTP server entry point
│   └── seed/main.go        # Database seeding command
├── app/
│   ├── api/                # HTTP response and error handling
│   │   ├── errors.go       # Centralized error mapping
│   │   ├── response.go     # JSON response helpers
│   │   └── response_test.go
│   ├── catalog/            # Catalog HTTP handlers
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── categories/         # Categories HTTP handlers
│   │   ├── handler.go
│   │   └── handler_test.go
│   ├── database/           # Database connection
│   │   └── pg.go
│   ├── logger/             # Structured logging
│   │   └── logger.go
│   ├── middleware/         # HTTP middlewares
│   │   ├── logger.go       # Request logging
│   │   ├── recovery.go     # Panic recovery
│   │   └── request_id.go   # Request ID generation
│   └── services/           # Business logic layer
│       ├── errors.go       # Domain errors
│       ├── catalog_service.go
│       ├── catalog_service_test.go
│       ├── categories_service.go
│       └── categories_service_test.go
├── models/                 # Data models and repositories
│   ├── category.go
│   ├── products.go
│   ├── variants.go
│   ├── products_repository.go
│   └── categories_repository.go
├── sql/                    # Database migrations
├── docs/                   # API documentation
│   ├── openapi.yaml        # OpenAPI 3.0 specification
│   └── README.md
├── test/e2e/               # End-to-end tests
└── .github/workflows/      # CI/CD pipelines
    └── go.yml              # GitHub Actions workflow
```

## Setup Code Repository

1. Create a github/bitbucket/gitlab repository and push all this code as-is.
2. Create a new branch, and provide a pull-request against the main branch with your changes. Instructions to follow.

## Application Setup

- Ensure you have Go installed on your machine.
- Ensure you have Docker installed on your machine.
- Important makefile targets:
  - `make help`: Show all available commands
  - `make tidy`: Tidy and vendor Go modules (runs `go mod tidy && go mod vendor`)
  - `make docker-up`: Start the required infrastructure services via docker containers
  - `make seed`: ⚠️ Will destroy and re-create the database tables
- `make test`: Run unit tests with coverage (excludes e2e)
  - `make test-unit`: Run only unit tests (fast, no database required)
  - `make test-e2e`: Run only end-to-end tests (requires PostgreSQL)
  - `make test-all`: Run unit + e2e tests sequentially
  - `make run`: Start the application
  - `make docker-down`: Stop the docker containers

## API Endpoints

> **Note:** All endpoints use the `/v1/` prefix for API versioning. Legacy routes without `/v1/` are also available for compatibility.

### Catalog

#### `GET /v1/catalog`
List all products with pagination and category information.

**Query Parameters:**
- `offset` (optional): Number of items to skip. Default: 0
- `limit` (optional): Maximum number of items to return. Default: 10, Min: 1, Max: 100

**Response:** `200 OK`
```json
{
  "products": [
    {
      "code": "PROD001",
      "price": 10.99,
      "category": {
        "code": "CLOTHING",
        "name": "Clothing"
      }
    }
  ],
  "total": 8
}
```

**Examples:**
```bash
# Get first page (default pagination)
curl http://localhost:8080/v1/catalog

# Get specific page
curl "http://localhost:8080/v1/catalog?offset=10&limit=5"

# Get with maximum items
curl "http://localhost:8080/v1/catalog?limit=100"
```

#### `GET /v1/catalog/{code}`
Get detailed information about a specific product including variants.

**Path Parameters:**
- `code`: Product code (e.g., "PROD001")

**Response:** `200 OK`
```json
{
  "code": "PROD001",
  "price": 10.99,
  "category": {
    "code": "CLOTHING",
    "name": "Clothing"
  },
  "variants": [
    {
      "name": "Variant A",
      "sku": "SKU001A",
      "price": 11.99
    },
    {
      "name": "Variant B",
      "sku": "SKU001B",
      "price": 10.99
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request`: Product code is required
- `404 Not Found`: Product does not exist
- `500 Internal Server Error`: Database error

**Notes:**
- Variants without a specific price inherit the product's base price

**Example:**
```bash
curl http://localhost:8080/v1/catalog/PROD001
```

### Categories

#### `GET /v1/categories`
List all available product categories.

**Response:** `200 OK`
```json
[
  {
    "code": "CLOTHING",
    "name": "Clothing"
  },
  {
    "code": "SHOES",
    "name": "Shoes"
  },
  {
    "code": "ACCESSORIES",
    "name": "Accessories"
  }
]
```

**Example:**
```bash
curl http://localhost:8080/v1/categories
```

#### `POST /v1/categories`
Create a new product category.

**Request Body:**
```json
{
  "code": "ELECTRONICS",
  "name": "Electronics"
}
```

**Response:** `201 Created`
```json
{
  "code": "ELECTRONICS",
  "name": "Electronics"
}
```

**Validation:**
- `code` is required (unique identifier)
- `name` is required (display name)
- Returns `400 Bad Request` if validation fails

**Example:**
```bash
curl -X POST http://localhost:8080/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"code":"ELECTRONICS","name":"Electronics"}'
```

## Error Responses

All error responses follow a standardized JSON format:

```json
{
  "code": "error_code",
  "message": "Human-readable error message"
}
```

### Error Codes

| Code | HTTP Status | Description | Example Message |
|------|-------------|-------------|-----------------|
| `invalid_input` | 400 | Invalid request parameters or body | `"offset must be a non-negative integer"` |
| `not_found` | 404 | Resource not found | `"Resource not found"` |
| `internal_error` | 500 | Internal server error | `"An internal error occurred"` |

### Specific Validation Messages

The API provides detailed validation feedback for common input errors:
- `"offset must be a non-negative integer"` - when offset parameter is negative or invalid
- `"limit must be a positive integer"` - when limit parameter is invalid
- `"priceLessThan must be a valid decimal number"` - when price filter is not a valid number
- `"priceLessThan must be a non-negative value"` - when price filter is negative
- `"category code and name are required"` - when creating a category with missing fields

### Request Tracing

All responses include an `X-Request-ID` header for distributed tracing:
```bash
curl -v http://localhost:8080/v1/catalog
# < X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

You can also provide your own request ID:
```bash
curl -H "X-Request-ID: my-custom-id" http://localhost:8080/v1/catalog
```

## Testing

The project includes comprehensive test coverage:

### Unit Tests (40 tests)
Fast, isolated tests using mocks. No database required.
```bash
make test-unit
```

| Package | Tests | Description |
|---------|-------|-------------|
| `app/api` | 2 | HTTP response helpers |
| `app/catalog` | 10 | Catalog handler tests |
| `app/categories` | 7 | Categories handler tests |
| `app/services` | 21 | Business logic tests |

### End-to-End Tests (22 scenarios)
Integration tests using real database and HTTP server.
```bash
# Ensure PostgreSQL is running
make docker-up

# Run e2e tests
make test-e2e
```

See [test/e2e/README.md](test/e2e/README.md) for detailed documentation.

## Architecture

### Layered Architecture

The application follows a clean layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer                              │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ catalog/handler │  │ categories/     │  ← HTTP concerns  │
│  │                 │  │ handler         │    only           │
│  └────────┬────────┘  └────────┬────────┘                   │
└───────────│─────────────────────│────────────────────────────┘
            │                     │
            ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ CatalogService  │  │ CategoriesService│ ← Business logic │
│  │ - ParsePagination│ │ - ListCategories │                  │
│  │ - ListProducts  │  │ - CreateCategory │                  │
│  │ - GetProductBy  │  └─────────┬────────┘                  │
│  │   Code          │            │                            │
│  └────────┬────────┘            │                            │
└───────────│─────────────────────│────────────────────────────┘
            │                     │
            ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐                   │
│  │ ProductsRepo    │  │ CategoriesRepo  │  ← Data access    │
│  └─────────────────┘  └─────────────────┘                   │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **Separation of Concerns**
   - **Handlers**: HTTP request/response handling only
   - **Services**: Business logic, validation, data transformation
   - **Repositories**: Database operations

2. **Interface-Based Design**
   - Interfaces defined where they are used (Go idiom)
   - Enables easy testing with mocks
   - Dependency injection via constructors

3. **Error Handling**
   - Specific validation errors with descriptive messages (`ErrInvalidOffset`, `ErrInvalidLimit`, etc.)
   - Domain errors (`ErrNotFound`, `ErrInvalidInput`) in services
   - Centralized error mapping via `api.ErrorHandler()` + `api.HandleError()`
   - Standardized error format: `{"code": "error_code", "message": "descriptive message"}`
   - Clear distinction between client errors (4xx) and server errors (5xx)

4. **Consistent API Responses**
   - Success responses via `api.OKResponse()` and `api.CreatedResponse()`
   - Error responses via `api.HandleError()` with format `{"code": "invalid_input", "message": "..."}`
   - Structured logging with request tracing (X-Request-ID header)

### Database
- PostgreSQL with GORM ORM
- Migration scripts in `sql/` directory
- Automatic table creation with `AutoMigrate`

### Models
- **Category**: Product categories (Clothing, Shoes, Accessories)
- **Product**: Products with code, price, and category relationship
- **Variant**: Product variants with optional custom pricing

## Project Status

✅ All requirements from [ASSIGNMENT.md](ASSIGNMENT.md) completed:
- ✅ Repository refactored to use interfaces
- ✅ Category model created and linked to products
- ✅ Catalog endpoint includes categories
- ✅ Pagination implemented (offset/limit)
- ✅ Product details endpoint with variants
- ✅ Categories endpoints (GET and POST)
- ✅ Comprehensive unit tests (40 tests)
- ✅ End-to-end tests (22 scenarios)
- ✅ Response helper functions
- ✅ Business logic in service layer (not handlers)
- ✅ Proper error handling (404 vs 500)
- ✅ Export comments on public symbols

Follow up for the assignment here: [ASSIGNMENT.md](ASSIGNMENT.md)
