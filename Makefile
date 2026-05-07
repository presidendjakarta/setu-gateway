.PHONY: help build run test clean setup migrate docker-up docker-down

# Default target
help:
	@echo "===================================="
	@echo "Setu API Gateway - Make Commands"
	@echo "===================================="
	@echo ""
	@echo "Available commands:"
	@echo "  make setup      - Setup dependencies and database"
	@echo "  make build      - Build the gateway binary"
	@echo "  make run        - Run the gateway"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make migrate    - Run database migrations"
	@echo "  make docker-up  - Start Docker Compose services"
	@echo "  make docker-down - Stop Docker Compose services"
	@echo "  make dev        - Run with live reload (requires air)"
	@echo ""

# Setup dependencies and database
setup:
	@echo "Setting up Setu Gateway..."
	@echo ""
	@echo "[1/3] Downloading Go dependencies..."
	go mod download
	@echo ""
	@echo "[2/3] Checking PostgreSQL..."
	@psql -h localhost -U postgres -d setu_gateway -c "SELECT 1" > /dev/null 2>&1 || \
		(echo "Creating database..." && \
		psql -h localhost -U postgres -c "CREATE DATABASE setu_gateway;" 2>/dev/null || \
		echo "Database may already exist or needs manual creation")
	@echo ""
	@echo "[3/3] Running migrations..."
	$(MAKE) migrate
	@echo ""
	@echo "Setup complete! Run 'make build' to build the gateway."

# Build the gateway
build:
	@echo "Building Setu Gateway..."
	go build -ldflags="-w -s" -o setu-gateway ./cmd/gateway
	@echo "Build complete: setu-gateway"

# Build for Windows
build-windows:
	@echo "Building Setu Gateway for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o setu-gateway.exe ./cmd/gateway
	@echo "Build complete: setu-gateway.exe"

# Run the gateway
run: build
	@echo "Starting Setu Gateway..."
	./setu-gateway

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run specific test
test-specific:
	@echo "Running specific test..."
	go test -v -run $(TEST) ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -benchtime=5s ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f setu-gateway setu-gateway.exe
	rm -f coverage.out coverage.html
	rm -rf tmp/
	@echo "Clean complete."

# Run database migrations
migrate:
	@echo "Running database migrations..."
	PGPASSWORD=postgres psql -h localhost -U postgres -d setu_gateway -f migrations/001_initial.up.sql
	@echo "Migrations complete."

# Start Docker Compose services
docker-up:
	@echo "Starting Docker Compose services..."
	docker-compose up -d
	@echo ""
	@echo "Services started:"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - Gateway: localhost:8080"
	@echo "  - Admin: localhost:8081"
	@echo "  - Metrics: localhost:9090"
	@echo "  - Prometheus: localhost:9091"
	@echo "  - Jaeger: localhost:16686"
	@echo ""
	@echo "View logs: docker-compose logs -f"

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down
	@echo "Services stopped."

# Restart Docker Compose services
docker-restart: docker-down docker-up

# View Docker logs
docker-logs:
	docker-compose logs -f $(SERVICE)

# Development mode with live reload (requires air)
dev:
	@echo "Starting development mode with live reload..."
	@if ! command -v air &> /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	air

# Debug mode with air (verbose)
dev-debug:
	@echo "Starting development mode with debug logging..."
	@if ! command -v air &> /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	air -d

# Install air for live reload
install-air:
	@echo "Installing air..."
	go install github.com/air-verse/air@latest
	@echo "Air installed."

# Check dependencies
check-deps:
	@echo "Checking dependencies..."
	@echo ""
	@echo "Go:"
	@go version
	@echo ""
	@echo "PostgreSQL:"
	@psql --version 2>/dev/null || echo "  Not installed"
	@echo ""
	@echo "Redis:"
	@redis-cli --version 2>/dev/null || echo "  Not installed"
	@echo ""
	@echo "Docker:"
	@docker --version 2>/dev/null || echo "  Not installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Security audit
audit:
	@echo "Running security audit..."
	gosec ./...

# Generate API documentation
docs:
	@echo "Generating documentation..."
	swag init -g cmd/gateway/main.go -o docs

# Database commands
db-status:
	@echo "Database status..."
	PGPASSWORD=postgres psql -h localhost -U postgres -d setu_gateway -c "\dt"

db-shell:
	@echo "Opening database shell..."
	PGPASSWORD=postgres psql -h localhost -U postgres -d setu_gateway

db-reset:
	@echo "WARNING: This will drop and recreate the database!"
	@read -p "Are you sure? (y/N) " confirm && [ "$$confirm" = "y" ] || exit 1
	PGPASSWORD=postgres psql -h localhost -U postgres -c "DROP DATABASE setu_gateway;"
	PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE setu_gateway;"
	$(MAKE) migrate

# Load test
load-test:
	@echo "Running load test..."
	@if ! command -v wrk &> /dev/null; then \
		echo "wrk not installed. Install it first."; \
		exit 1; \
	fi
	wrk -t12 -c400 -d30s http://localhost:8080/health

# Performance profiling
profile:
	@echo "Starting profiler..."
	@echo "Visit http://localhost:6060/debug/pprof"
	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Create release
release: clean test
	@echo "Creating release..."
	@VERSION=$$(grep "version:" configs/gateway.yaml | awk '{print $$2}')
	@echo "Version: $$VERSION"
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dist/setu-gateway-$$VERSION-linux-amd64 ./cmd/gateway
	GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o dist/setu-gateway-$$VERSION-darwin-amd64 ./cmd/gateway
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/setu-gateway-$$VERSION-windows-amd64.exe ./cmd/gateway
	@echo "Release builds created in dist/"
