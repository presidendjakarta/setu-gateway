@echo off
REM Setup script for Windows - Sets up PostgreSQL and Redis locally

echo ====================================
echo Setu Gateway - Windows Setup
echo ====================================
echo.

REM Check if Docker is available (optional)
where docker >nul 2>nul
if %errorlevel% equ 0 (
    echo [INFO] Docker detected. You can use docker-compose if preferred.
    echo.
)

REM Check if PostgreSQL is running
echo [1/4] Checking PostgreSQL...
psql --version >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARN] PostgreSQL not found in PATH.
    echo.
    echo Please install PostgreSQL 16+:
    echo https://www.postgresql.org/download/windows/
    echo.
    echo Or use Docker:
    echo   docker run -d --name setu-db -e POSTGRES_DB=setu_gateway -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:16-alpine
    echo.
    pause
    exit /b 1
)

REM Check database connection
echo [INFO] Testing database connection...
set PGPASSWORD=postgres
psql -h localhost -U postgres -d setu_gateway -c "SELECT 1" >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARN] Cannot connect to database 'setu_gateway'.
    echo.
    echo Creating database...
    psql -h localhost -U postgres -c "CREATE DATABASE setu_gateway;" 2>nul
    if %errorlevel% equ 0 (
        echo [OK] Database created successfully.
    ) else (
        echo [ERROR] Failed to create database. Please check PostgreSQL is running.
        pause
        exit /b 1
    )
) else (
    echo [OK] Database connection successful.
)

REM Check if Redis is available (optional)
echo.
echo [2/4] Checking Redis (optional)...
redis-cli ping >nul 2>nul
if %errorlevel% equ 0 (
    echo [OK] Redis is running.
) else (
    echo [INFO] Redis not detected. Rate limiting will use in-memory store.
    echo.
    echo To install Redis:
    echo https://redis.io/download/
    echo.
    echo Or use Docker:
    echo   docker run -d --name setu-redis -p 6379:6379 redis:7-alpine
)

REM Run migrations
echo.
echo [3/4] Running database migrations...
set PGPASSWORD=postgres
psql -h localhost -U postgres -d setu_gateway -f migrations\001_initial.up.sql
if %errorlevel% equ 0 (
    echo [OK] Migrations completed successfully.
) else (
    echo [ERROR] Migration failed.
    pause
    exit /b 1
)

REM Build the gateway
echo.
echo [4/4] Building Setu Gateway...
go build -o setu-gateway.exe ./cmd/gateway
if %errorlevel% equ 0 (
    echo [OK] Build successful!
) else (
    echo [ERROR] Build failed.
    pause
    exit /b 1
)

echo.
echo ====================================
echo Setup Complete!
echo ====================================
echo.
echo To start the gateway:
echo   setu-gateway.exe
echo.
echo Gateway will be available at:
echo   - Gateway: http://localhost:8080
echo   - Admin:   http://localhost:8081
echo   - Metrics: http://localhost:9090
echo.
echo To test:
echo   curl http://localhost:8080/health
echo.
pause
