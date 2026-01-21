@echo off
setlocal enabledelayedexpansion

:: ============================================================================
:: CloudCompare + PoissonRecon Batch Processing Script for LAS Files (Windows)
:: ============================================================================
:: This script processes LAS point cloud files through:
:: 1. CloudCompare CLI - Import LAS, compute normals, convert to DIP, export PLY
:: 2. PoissonRecon - Poisson Surface Reconstruction with density scalar field
:: 3. CloudCompare CLI - Import point cloud and mesh into project for manual
::                       filtering by density SF value
::
:: The density scalar field from PoissonRecon allows you to filter out
:: low-confidence areas of the reconstruction in CloudCompare by adjusting
:: the SF display min value and exporting the filtered mesh.
::
:: Prerequisites:
:: - CloudCompare (with CLI support)
:: - PoissonRecon from https://github.com/mkazhdan/PoissonRecon
::
:: Poisson Surface Reconstruction Parameters:
:: - Octree depth = 11
:: - Boundary = Neumann
:: - Samples per node = 1.5
:: - Point weight = 2.0
:: - Threads = 16
:: - Output density as SF = Yes (for manual filtering in CloudCompare)
:: ============================================================================

:: Configuration
:: ----------------------------------------------------------------------------
:: Path to CloudCompare executable
if not defined CLOUDCOMPARE_PATH (
    set "CLOUDCOMPARE_PATH=C:\Program Files\CloudCompare\CloudCompare.exe"
)

:: Path to PoissonRecon executable
if not defined POISSONRECON_PATH (
    set "POISSONRECON_PATH=PoissonRecon.exe"
)

:: Input directory containing LAS files (first argument or current directory)
set "INPUT_DIR=%~1"
if "%INPUT_DIR%"=="" set "INPUT_DIR=%CD%"

:: Output directory for processed files (relative to input)
set "OUTPUT_DIR=Processed"

:: Normal Computation Parameters
set "OCTREE_RADIUS=AUTO"
set "KNN=6"

:: Poisson Reconstruction Parameters
set "OCTREE_DEPTH=11"
set "SAMPLES_PER_NODE=1.5"
set "POINT_WEIGHT=2.0"
set "THREADS=16"
set "BOUNDARY_TYPE=3"

:: Counters
set "SUCCESS_COUNT=0"
set "FAIL_COUNT=0"
set "TOTAL_COUNT=0"

:: Temporary directory
set "TEMP_DIR=%TEMP%\cc_poisson_%RANDOM%"

:: ============================================================================
:: Main Script
:: ============================================================================

echo.
echo ============================================================================
echo  CloudCompare + PoissonRecon Batch Processing Script for Windows
echo ============================================================================
echo.

:: Check for help flag
if "%~1"=="-h" goto :show_help
if "%~1"=="--help" goto :show_help
if "%~1"=="/?" goto :show_help

:: Check if CloudCompare exists
call :check_cloudcompare
if %ERRORLEVEL% neq 0 exit /b 1

:: Check if PoissonRecon exists
call :check_poissonrecon
if %ERRORLEVEL% neq 0 exit /b 1

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

:: Create temporary directory
mkdir "%TEMP_DIR%" 2>nul
echo [INFO] Temporary directory: %TEMP_DIR%

:: Count LAS files
for %%f in (*.las *.LAS) do (
    set /a TOTAL_COUNT+=1
)

if %TOTAL_COUNT%==0 (
    echo [ERROR] No LAS files found in: %INPUT_DIR%
    goto :cleanup
)

echo [INFO] Found %TOTAL_COUNT% LAS file(s) to process
echo.
echo [INFO] Processing Parameters:
echo [INFO]   Normal Computation:
echo [INFO]     - Octree Radius: %OCTREE_RADIUS%
echo [INFO]     - MST Orientation KNN: %KNN%
echo [INFO]   Poisson Reconstruction:
echo [INFO]     - Octree Depth: %OCTREE_DEPTH%
echo [INFO]     - Samples per Node: %SAMPLES_PER_NODE%
echo [INFO]     - Point Weight: %POINT_WEIGHT%
echo [INFO]     - Boundary: Neumann
echo [INFO]     - Threads: %THREADS%
echo [INFO]     - Output Density as SF: YES
echo [INFO]     - Interpolate Colors: YES
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

echo.
echo [INFO] Output files are in: %CD%\%OUTPUT_DIR%
echo [INFO]   - [filename].bin : CloudCompare project with point cloud and mesh
echo.
echo [INFO] The mesh includes a 'Density' scalar field from PoissonRecon.
echo [INFO] Open the .bin file in CloudCompare to filter out low-density areas:
echo [INFO]   1. Select mesh ^> Properties ^> SF Display ^> adjust min value
echo [INFO]   2. Edit ^> Scalar Fields ^> Filter by Value
echo [INFO]   3. Export the filtered mesh

:cleanup
:: Clean up temporary directory
if exist "%TEMP_DIR%" (
    echo [INFO] Cleaning up temporary files...
    rmdir /s /q "%TEMP_DIR%" 2>nul
)

popd
exit /b %FAIL_COUNT%

:: ============================================================================
:: Functions
:: ============================================================================

:check_cloudcompare
if exist "%CLOUDCOMPARE_PATH%" (
    echo [INFO] Using CloudCompare: %CLOUDCOMPARE_PATH%
    exit /b 0
)

:: Try alternative paths
if exist "C:\Program Files\CloudCompare\CloudCompare.exe" (
    set "CLOUDCOMPARE_PATH=C:\Program Files\CloudCompare\CloudCompare.exe"
    echo [INFO] Found CloudCompare at default location.
    exit /b 0
)
if exist "C:\Program Files (x86)\CloudCompare\CloudCompare.exe" (
    set "CLOUDCOMPARE_PATH=C:\Program Files (x86)\CloudCompare\CloudCompare.exe"
    echo [INFO] Found CloudCompare at x86 location.
    exit /b 0
)

echo [ERROR] CloudCompare not found!
echo [ERROR] Please set CLOUDCOMPARE_PATH environment variable or install CloudCompare.
exit /b 1

:check_poissonrecon
:: Check if PoissonRecon is in PATH or at specified location
where "%POISSONRECON_PATH%" >nul 2>&1
if %ERRORLEVEL%==0 (
    echo [INFO] Using PoissonRecon: %POISSONRECON_PATH%
    exit /b 0
)

if exist "%POISSONRECON_PATH%" (
    echo [INFO] Using PoissonRecon: %POISSONRECON_PATH%
    exit /b 0
)

:: Try common locations
if exist "C:\PoissonRecon\PoissonRecon.exe" (
    set "POISSONRECON_PATH=C:\PoissonRecon\PoissonRecon.exe"
    echo [INFO] Found PoissonRecon at: %POISSONRECON_PATH%
    exit /b 0
)
if exist "%USERPROFILE%\PoissonRecon\PoissonRecon.exe" (
    set "POISSONRECON_PATH=%USERPROFILE%\PoissonRecon\PoissonRecon.exe"
    echo [INFO] Found PoissonRecon at: %POISSONRECON_PATH%
    exit /b 0
)

echo [ERROR] PoissonRecon not found!
echo [ERROR] Please install PoissonRecon from https://github.com/mkazhdan/PoissonRecon
echo [ERROR] Or set POISSONRECON_PATH environment variable.
exit /b 1

:process_file
set "INPUT_FILE=%~1"
set "FILENAME=%~n1"
set "PLY_WITH_NORMALS=%TEMP_DIR%\%FILENAME%_normals.ply"
set "MESH_PLY=%TEMP_DIR%\%FILENAME%_mesh.ply"
set "OUTPUT_BIN=%OUTPUT_DIR%\%FILENAME%.bin"

echo.
echo ============================================================================
echo  Processing: %FILENAME%
echo ============================================================================
echo [INFO] Input:  %INPUT_FILE%
echo [INFO] Output: %OUTPUT_BIN%
echo.

:: -------------------------------------------------------------------------
:: Step 1: CloudCompare - Load LAS, compute normals, convert to DIP, export PLY
:: -------------------------------------------------------------------------
echo [INFO] Step 1: Computing normals and converting to DIP with CloudCompare...
echo [INFO]   - Octree Radius: %OCTREE_RADIUS%
echo [INFO]   - MST Orientation KNN: %KNN%

"%CLOUDCOMPARE_PATH%" -SILENT -AUTO_SAVE OFF ^
    -O "%INPUT_FILE%" ^
    -OCTREE_NORMALS %OCTREE_RADIUS% ^
    -ORIENT_NORMS_MST %KNN% ^
    -NORMALS_TO_DIP ^
    -C_EXPORT_FMT PLY -PLY_EXPORT_FMT ASCII ^
    -SAVE_CLOUDS FILE "%PLY_WITH_NORMALS%"

if not exist "%PLY_WITH_NORMALS%" (
    echo [ERROR] Failed to generate PLY with normals: %PLY_WITH_NORMALS%
    set /a FAIL_COUNT+=1
    goto :eof
)
echo [SUCCESS] Generated PLY with normals

:: -------------------------------------------------------------------------
:: Step 2: PoissonRecon - Poisson Surface Reconstruction with density SF
:: -------------------------------------------------------------------------
echo.
echo [INFO] Step 2: Running Poisson Surface Reconstruction...
echo [INFO]   - Octree depth: %OCTREE_DEPTH%
echo [INFO]   - Samples per node: %SAMPLES_PER_NODE%
echo [INFO]   - Point weight: %POINT_WEIGHT%
echo [INFO]   - Boundary type: %BOUNDARY_TYPE% (Neumann)
echo [INFO]   - Threads: %THREADS%
echo [INFO]   - Output density as SF: YES (for filtering in CloudCompare)

"%POISSONRECON_PATH%" ^
    --in "%PLY_WITH_NORMALS%" ^
    --out "%MESH_PLY%" ^
    --depth %OCTREE_DEPTH% ^
    --samplesPerNode %SAMPLES_PER_NODE% ^
    --pointWeight %POINT_WEIGHT% ^
    --bType %BOUNDARY_TYPE% ^
    --threads %THREADS% ^
    --density ^
    --colors

if not exist "%MESH_PLY%" (
    echo [ERROR] Failed to generate mesh: %MESH_PLY%
    set /a FAIL_COUNT+=1
    goto :eof
)
echo [SUCCESS] Generated mesh with density scalar field

:: -------------------------------------------------------------------------
:: Step 3: CloudCompare - Import point cloud and mesh, save as .bin project
:: -------------------------------------------------------------------------
echo.
echo [INFO] Step 3: Creating CloudCompare project (.bin) for manual filtering...
echo [INFO]   The mesh contains a 'Density' scalar field from PoissonRecon.
echo [INFO]   In CloudCompare, you can:
echo [INFO]     1. Select the mesh
echo [INFO]     2. Display the Density SF (Edit ^> Scalar Fields ^> Set Active SF)
echo [INFO]     3. Adjust SF display range to visualize low-density areas
echo [INFO]     4. Filter by SF value (Edit ^> Scalar Fields ^> Filter by Value)
echo [INFO]     5. Export the filtered mesh

"%CLOUDCOMPARE_PATH%" -SILENT -AUTO_SAVE OFF ^
    -O "%PLY_WITH_NORMALS%" ^
    -O "%MESH_PLY%" ^
    -SAVE_CLOUDS ALL FILE "%OUTPUT_BIN%"

if exist "%OUTPUT_BIN%" (
    echo [SUCCESS] Created CloudCompare project: %OUTPUT_BIN%
    echo.
    echo [INFO] Next steps in CloudCompare GUI:
    echo [INFO]   1. Open %OUTPUT_BIN%
    echo [INFO]   2. Select the mesh in DB Tree
    echo [INFO]   3. Properties panel ^> SF Display ^> adjust 'displayed' min value
    echo [INFO]   4. Use 'Edit ^> Scalar Fields ^> Filter by Value' to remove low-density vertices
    echo [INFO]   5. Export filtered mesh via 'File ^> Save'
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] Failed to create project: %OUTPUT_BIN%
    set /a FAIL_COUNT+=1
)

goto :eof

:: ============================================================================
:: Help
:: ============================================================================

:show_help
echo Usage: %~nx0 [INPUT_DIRECTORY]
echo.
echo Process LAS point cloud files with CloudCompare and PoissonRecon.
echo The output mesh includes a density scalar field for manual filtering.
echo.
echo Arguments:
echo   INPUT_DIRECTORY   Directory containing LAS files (default: current directory)
echo.
echo Environment Variables:
echo   CLOUDCOMPARE_PATH   Path to CloudCompare executable
echo   POISSONRECON_PATH   Path to PoissonRecon executable
echo.
echo Output:
echo   Processed files will be saved to INPUT_DIRECTORY\Processed\
echo   Each .bin file contains the point cloud and mesh with density SF.
echo.
echo Processing Pipeline:
echo   1. Import LAS file (CloudCompare)
echo   2. Compute Normals using Octree (radius=AUTO)
echo   3. Orient Normals using MST (knn=6)
echo   4. Convert Normals to Dip/Dip Direction
echo   5. Export as PLY with normals
echo   6. Poisson Surface Reconstruction (PoissonRecon)
echo      - Octree depth: 11
echo      - Samples per node: 1.5
echo      - Point weight: 2.0
echo      - Boundary: Neumann
echo      - Output density as SF: YES
echo      - Interpolate colors: YES
echo      - Threads: 16
echo   7. Save CloudCompare project (.bin) with point cloud and mesh
echo.
echo Manual Filtering in CloudCompare:
echo   1. Open the .bin project file
echo   2. Select the mesh in the DB Tree
echo   3. In Properties panel, adjust SF Display min value
echo   4. Use 'Edit ^> Scalar Fields ^> Filter by Value' to remove vertices
echo   5. Export the filtered mesh
echo.
echo Prerequisites:
echo   - CloudCompare: https://www.cloudcompare.org/
echo   - PoissonRecon: https://github.com/mkazhdan/PoissonRecon
echo.
echo Example:
echo   %~nx0 C:\PointClouds
echo   set POISSONRECON_PATH=D:\Tools\PoissonRecon.exe ^&^& %~nx0
exit /b 0
