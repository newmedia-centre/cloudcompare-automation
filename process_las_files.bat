@echo off
setlocal enabledelayedexpansion

:: ============================================================================
:: CloudCompare Batch Processing Script for LAS Files (Windows)
:: ============================================================================
:: This script processes LAS point cloud files through CloudCompare CLI to:
:: 1. Import LAS files
:: 2. Compute normals (Triangulation + MST orientation knn=6)
:: 3. Convert normals to Dip/Dip Direction
:: 4. Perform Poisson Surface Reconstruction
:: 5. Save the project as .bin file
:: ============================================================================

:: Configuration
:: ----------------------------------------------------------------------------
:: Path to CloudCompare executable (adjust based on your installation)
:: Common paths:
::   - C:\Program Files\CloudCompare\CloudCompare.exe
::   - C:\Program Files (x86)\CloudCompare\CloudCompare.exe

if not defined CLOUDCOMPARE_PATH (
    set "CLOUDCOMPARE_PATH=C:\Program Files\CloudCompare\CloudCompare.exe"
)

:: Input directory containing LAS files (first argument or current directory)
set "INPUT_DIR=%~1"
if "%INPUT_DIR%"=="" set "INPUT_DIR=%CD%"

:: Output directory for processed files (relative to input)
set "OUTPUT_DIR=Processed"

:: Poisson Reconstruction Parameters
set "OCTREE_DEPTH=11"
set "SAMPLES_PER_NODE=1.5"
set "POINT_WEIGHT=2.0"
set "THREADS=16"
set "BOUNDARY=NEUMANN"

:: Normal Computation Parameters
set "KNN=6"

:: Counters
set "SUCCESS_COUNT=0"
set "FAIL_COUNT=0"
set "TOTAL_COUNT=0"

:: ============================================================================
:: Main Script
:: ============================================================================

echo.
echo ============================================================================
echo  CloudCompare Batch Processing Script for Windows
echo ============================================================================
echo.

:: Check for help flag
if "%~1"=="-h" goto :show_help
if "%~1"=="--help" goto :show_help
if "%~1"=="/?" goto :show_help

:: Check if CloudCompare exists
if not exist "%CLOUDCOMPARE_PATH%" (
    echo [ERROR] CloudCompare not found at: %CLOUDCOMPARE_PATH%
    echo.
    echo Please set the CLOUDCOMPARE_PATH environment variable or edit this script.
    echo Example: set CLOUDCOMPARE_PATH=C:\Program Files\CloudCompare\CloudCompare.exe
    echo.

    :: Try alternative paths
    if exist "C:\Program Files\CloudCompare\CloudCompare.exe" (
        set "CLOUDCOMPARE_PATH=C:\Program Files\CloudCompare\CloudCompare.exe"
        echo [INFO] Found CloudCompare at default location.
    ) else if exist "C:\Program Files (x86)\CloudCompare\CloudCompare.exe" (
        set "CLOUDCOMPARE_PATH=C:\Program Files (x86)\CloudCompare\CloudCompare.exe"
        echo [INFO] Found CloudCompare at x86 location.
    ) else (
        exit /b 1
    )
)

echo [INFO] Using CloudCompare: %CLOUDCOMPARE_PATH%

:: Validate input directory
if not exist "%INPUT_DIR%" (
    echo [ERROR] Input directory does not exist: %INPUT_DIR%
    exit /b 1
)

:: Change to input directory
pushd "%INPUT_DIR%"
echo [INFO] Working directory: %CD%

:: Create output directory
if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%"
)
echo [INFO] Output directory: %OUTPUT_DIR%

:: Count LAS files
for %%f in (*.las *.LAS) do (
    set /a TOTAL_COUNT+=1
)

if %TOTAL_COUNT%==0 (
    echo [ERROR] No LAS files found in: %INPUT_DIR%
    popd
    exit /b 1
)

echo [INFO] Found %TOTAL_COUNT% LAS file(s) to process
echo.

:: Process each LAS file
for %%f in (*.las *.LAS) do (
    call :process_file "%%f"
)

:: Summary
echo.
echo ============================================================================
echo  Processing Complete
echo ============================================================================
echo [INFO] Successfully processed: %SUCCESS_COUNT% file(s)
if %FAIL_COUNT% GTR 0 (
    echo [ERROR] Failed to process: %FAIL_COUNT% file(s)
)

popd
exit /b %FAIL_COUNT%

:: ============================================================================
:: Functions
:: ============================================================================

:process_file
set "INPUT_FILE=%~1"
set "FILENAME=%~n1"
set "OUTPUT_FILE=%OUTPUT_DIR%\%FILENAME%.bin"

echo.
echo ============================================================================
echo  Processing: %FILENAME%
echo ============================================================================
echo [INFO] Input:  %INPUT_FILE%
echo [INFO] Output: %OUTPUT_FILE%
echo.

:: Build and execute the CloudCompare command
:: Note: CloudCompare CLI processes commands in sequence
::
:: Command breakdown:
:: -SILENT              : Suppress GUI dialogs
:: -AUTO_SAVE OFF       : Disable auto-save prompts
:: -O                   : Open/load file
:: -COMPUTE_NORMALS     : Compute normals for the point cloud
:: -NORMALS_TO_DIP      : Convert normals to Dip/Dip Direction
:: -POISSON             : Poisson Surface Reconstruction
:: -C_EXPORT_FMT BIN    : Set cloud export format to BIN
:: -M_EXPORT_FMT BIN    : Set mesh export format to BIN
:: -SAVE_CLOUDS         : Save point clouds

echo [INFO] Executing CloudCompare CLI...
echo.

"%CLOUDCOMPARE_PATH%" -SILENT -AUTO_SAVE OFF ^
    -O "%INPUT_FILE%" ^
    -COMPUTE_NORMALS ^
    -NORMALS_TO_DIP ^
    -POISSON %OCTREE_DEPTH% %SAMPLES_PER_NODE% %BOUNDARY% ^
    -C_EXPORT_FMT BIN ^
    -M_EXPORT_FMT BIN ^
    -SAVE_CLOUDS FILE "%OUTPUT_FILE%"

if %ERRORLEVEL%==0 (
    echo [SUCCESS] Successfully processed: %FILENAME%
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] Failed to process: %FILENAME% (exit code: %ERRORLEVEL%)
    set /a FAIL_COUNT+=1
)

goto :eof

:: ============================================================================
:: Help
:: ============================================================================

:show_help
echo Usage: %~nx0 [INPUT_DIRECTORY]
echo.
echo Process LAS point cloud files with CloudCompare.
echo.
echo Arguments:
echo   INPUT_DIRECTORY   Directory containing LAS files (default: current directory)
echo.
echo Environment Variables:
echo   CLOUDCOMPARE_PATH   Path to CloudCompare executable
echo.
echo Output:
echo   Processed files will be saved to INPUT_DIRECTORY\Processed\
echo.
echo Processing Steps:
echo   1. Import LAS file
echo   2. Compute Normals (Triangulation, MST orientation knn=6)
echo   3. Convert Normals to Dip/Dip Direction
echo   4. Poisson Surface Reconstruction (depth=%OCTREE_DEPTH%)
echo   5. Save as .bin project file
echo.
echo Example:
echo   %~nx0 C:\PointClouds
echo   set CLOUDCOMPARE_PATH=D:\CloudCompare\CloudCompare.exe ^&^& %~nx0
exit /b 0
