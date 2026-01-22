@echo off
REM ============================================================================
REM Build script for CloudCompare Automation TUI
REM ============================================================================

setlocal EnableDelayedExpansion

echo ============================================================================
echo CloudCompare Automation TUI - Build Script
echo ============================================================================
echo.

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go is not installed or not in PATH!
    echo         Please install Go from: https://go.dev/dl/
    exit /b 1
)

echo [INFO] Go version:
go version
echo.

REM Get dependencies
echo [STEP 1/3] Downloading dependencies...
go mod download
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to download dependencies!
    exit /b 1
)
echo [SUCCESS] Dependencies downloaded.
echo.

REM Tidy up modules
echo [STEP 2/3] Tidying modules...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo [WARNING] go mod tidy had issues, continuing anyway...
)
echo.

REM Build the application
echo [STEP 3/3] Building application...
set OUTPUT=cloudcompare-tui.exe
go build -o %OUTPUT% ./cmd/cloudcompare-tui
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Build failed!
    exit /b 1
)

echo.
echo ============================================================================
echo Build Complete!
echo ============================================================================
echo.
echo Output: %OUTPUT%
echo.
echo To run the TUI:
echo   .\%OUTPUT%
echo.
echo ============================================================================

endlocal
