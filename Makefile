.PHONY: help tidy seed run test test-unit test-e2e test-all test-ci docker-up docker-down lint

help ::
	@echo "Available commands:"
	@echo "  make tidy       - Tidy and vendor Go modules"
	@echo "  make seed       - Seed the database with test data"
	@echo "  make run        - Run the application server"
	@echo "  make test       - Run all tests with coverage"
	@echo "  make test-unit  - Run only unit tests (excludes e2e)"
	@echo "  make test-e2e   - Run only e2e tests (requires PostgreSQL)"
	@echo "  make test-all   - Run unit tests + e2e tests sequentially"
	@echo "  make test-ci    - Run tests in CI environment"
	@echo "  make lint       - Run linter"
	@echo "  make docker-up  - Start Docker containers"
	@echo "  make docker-down - Stop Docker containers"

tidy ::
	@go mod tidy && go mod vendor

seed ::
	@go run cmd/seed/main.go

run ::
	@go run cmd/server/main.go

test ::
	@go test -v -count=1 -race $$(go list ./... | grep -v /test/e2e) -coverprofile=coverage.out -covermode=atomic

test-unit ::
	@echo "Running unit tests..."
	@go test -v -count=1 -race $$(go list ./... | grep -v /test/e2e) -coverprofile=coverage.out -covermode=atomic

test-e2e ::
	@echo "Running e2e tests..."
	@echo "Make sure PostgreSQL is running and test database is configured"
	@go test -v -count=1 ./test/e2e/...

test-all ::
	@echo "Running all tests (unit + e2e)..."
	@$(MAKE) test-unit
	@$(MAKE) test-e2e

test-ci ::
	@echo "Running tests in CI environment..."
	@go test -v -race -coverprofile=coverage-unit.out -covermode=atomic $$(go list ./... | grep -v /test/e2e)
	@go test -v -coverprofile=coverage-e2e.out -covermode=atomic ./test/e2e/...

lint ::
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install with: brew install golangci-lint"; exit 1; }
	@golangci-lint run --timeout=5m

docker-up ::
	docker compose up -d

docker-down ::
	docker compose down
