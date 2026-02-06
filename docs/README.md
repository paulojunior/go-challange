# API Documentation

This directory contains the OpenAPI 3.0 specification for the Product Catalog API.

## Viewing the Documentation

### Option 1: Swagger UI (Online)

1. Go to [Swagger Editor](https://editor.swagger.io/)
2. Click `File` → `Import file`
3. Select `openapi.yaml` from this directory

### Option 2: Swagger UI (Local with Docker)

```bash
docker run -p 8081:8080 -e SWAGGER_JSON=/docs/openapi.yaml -v $(pwd)/docs:/docs swaggerapi/swagger-ui
```

Then open http://localhost:8081 in your browser.

### Option 3: Redoc (Local with npx)

```bash
npx @redocly/cli preview-docs docs/openapi.yaml
```

### Option 4: VS Code Extension

Install the [OpenAPI (Swagger) Editor](https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi) extension and open `openapi.yaml`.

## API Versioning

The API uses URL-based versioning with `/v1/` prefix:

```
GET /v1/catalog
GET /v1/categories
```

## Authentication

Currently, the API does not require authentication. This will be added in future versions.

## Request Tracing

All requests can include an `X-Request-ID` header for distributed tracing:

```bash
curl -H "X-Request-ID: 550e8400-e29b-41d4-a716-446655440000" \
  http://localhost:8080/v1/catalog
```

If not provided, the server will generate one automatically and include it in the response headers.

## Error Handling

All error responses follow a standardized format:

```json
{
  "code": "error_code",
  "message": "Human-readable error message"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_input` | 400 | Invalid request parameters or body |
| `not_found` | 404 | Resource not found |
| `internal_error` | 500 | Internal server error |

## Examples

### List Products with Filters

```bash
# Get first 10 products
curl http://localhost:8080/v1/catalog

# Get products from CLOTHING category with price < 50
curl "http://localhost:8080/v1/catalog?category=CLOTHING&priceLessThan=50"

# Pagination
curl "http://localhost:8080/v1/catalog?offset=10&limit=20"
```

### Get Product Details

```bash
curl http://localhost:8080/v1/catalog/PROD001
```

### List Categories

```bash
curl http://localhost:8080/v1/categories
```

### Create Category

```bash
curl -X POST http://localhost:8080/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"code": "SHOES", "name": "Shoes"}'
```

## Changelog

### Version 1.0.0 (Current)

- Initial API release
- Product catalog endpoints
- Category management endpoints
- Structured error responses
- Request tracing support
