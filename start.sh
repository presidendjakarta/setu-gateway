#!/bin/bash
# Start Setu Gateway locally

echo "===================================="
echo "Starting Setu API Gateway"
echo "===================================="
echo ""

# Check if binary exists
if [ ! -f "./setu-gateway" ]; then
    echo "[ERROR] setu-gateway binary not found."
    echo "Please run setup.sh first."
    exit 1
fi

# Check if PostgreSQL is running
echo "[INFO] Checking database connection..."
if ! PGPASSWORD=postgres psql -h localhost -U postgres -d setu_gateway -c "SELECT 1" &> /dev/null; then
    echo "[ERROR] Cannot connect to PostgreSQL."
    echo "Please start PostgreSQL first."
    exit 1
fi
echo "[OK] Database connected."

echo ""
echo "Starting gateway..."
echo ""

# Start the gateway
./setu-gateway
