@echo off
REM Start Setu Gateway locally

echo ====================================
echo Starting Setu API Gateway
echo ====================================
echo.

REM Check if binary exists
if not exist "setu-gateway.exe" (
    echo [ERROR] setu-gateway.exe not found.
    echo Please run setup-windows.bat first.
    pause
    exit /b 1
)

REM Check if PostgreSQL is running
echo [INFO] Checking database connection...
set PGPASSWORD=postgres
psql -h localhost -U postgres -d setu_gateway -c "SELECT 1" >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Cannot connect to PostgreSQL.
    echo Please start PostgreSQL first.
    echo.
    pause
    exit /b 1
)
echo [OK] Database connected.

echo.
echo Starting gateway...
echo.

REM Start the gateway
setu-gateway.exe
