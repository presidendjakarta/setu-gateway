@echo off
REM Start air live reload for development

echo ====================================
echo Starting Air Live Reload
echo ====================================
echo.

REM Check if air is installed
where air >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Air is not installed.
    echo Installing air...
    go install github.com/air-verse/air@latest
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to install air.
        pause
        exit /b 1
    )
)

REM Create tmp directory if not exists
if not exist "tmp" (
    mkdir tmp
)

echo [INFO] Starting air in debug mode...
echo.
echo Watching for file changes...
echo Press Ctrl+C to stop
echo.

REM Start air with debug mode
air -d
