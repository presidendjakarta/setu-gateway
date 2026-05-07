#!/bin/bash
# Setup script for Linux/Mac - Sets up PostgreSQL and Redis locally

echo "===================================="
echo "Setu Gateway - Setup"
echo "===================================="
echo ""

# Check if Docker is available (optional)
if command -v docker &> /dev/null; then
    echo "[INFO] Docker detected. You can use docker-compose if preferred."
    echo ""
fi

# Check if PostgreSQL is running
echo "[1/4] Checking PostgreSQL..."
if ! command -v psql &> /dev/null; then
    echo "[WARN] PostgreSQL not found in PATH."
    echo ""
    echo "Please install PostgreSQL 16+:"
    echo "  - Ubuntu/Debian: sudo apt install postgresql"
    echo "  - macOS: brew install postgresql"
    echo "  - Or use Docker:"
    echo "    docker run -d --name setu-db -e POSTGRES_DB=setu_gateway -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:16-alpine"
    echo ""
    exit 1
fi

# Check database connection
echo "[INFO] Testing database connection..."
if ! PGPASSWORD=postgres psql -h localhost -U postgres -d setu_gateway -c "SELECT 1" &> /dev/null; then
    echo "[WARN] Cannot connect to database 'setu_gateway'."
    echo ""
    echo "Creating database..."
    PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE setu_gateway;" 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "[OK] Database created successfully."
    else
        echo "[ERROR] Failed to create database. Please check PostgreSQL is running."
        exit 1
    fi
else
    echo "[OK] Database connection successful."
fi

# Check if Redis is available (optional)
echo ""
echo "[2/4] Checking Redis (optional)..."
if command -v redis-cli &> /dev/null && redis-cli ping &> /dev/null; then
    echo "[OK] Redis is running."
else
    echo "[INFO] Redis not detected. Rate limiting will use in-memory store."
    echo ""
    echo "To install Redis:"
    echo "  - Ubuntu/Debian: sudo apt install redis-server"
    echo "  - macOS: brew install redis"
    echo "  - Or use Docker:"
    echo "    docker run -d --name setu-redis -p 6379:6379 redis:7-alpine"
fi

# Run migrations
echo ""
echo "[3/4] Running database migrations..."
PGPASSWORD=postgres psql -h localhost -U postgres -d setu_gateway -f migrations/001_initial.up.sql
if [ $? -eq 0 ]; then
    echo "[OK] Migrations completed successfully."
else
    echo "[ERROR] Migration failed."
    exit 1
fi

# Build the gateway
echo ""
echo "[4/4] Building Setu Gateway..."
go build -o setu-gateway ./cmd/gateway
if [ $? -eq 0 ]; then
    echo "[OK] Build successful!"
else
    echo "[ERROR] Build failed."
    exit 1
fi

echo ""
echo "===================================="
echo "Setup Complete!"
echo "===================================="
echo ""
echo "To start the gateway:"
echo "  ./setu-gateway"
echo ""
echo "Gateway will be available at:"
echo "  - Gateway: http://localhost:8080"
echo "  - Admin:   http://localhost:8081"
echo "  - Metrics: http://localhost:9090"
echo ""
echo "To test:"
echo "  curl http://localhost:8080/health"
echo ""
