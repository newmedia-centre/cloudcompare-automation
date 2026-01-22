@echo off
REM ============================================================================
REM CloudComPy Conda Environment Setup Script for Windows
REM ============================================================================
REM
REM This script sets up the CloudComPy311 conda environment with all required
REM packages for CloudComPy (Python bindings for CloudCompare).
REM
REM Prerequisites:
REM   - Anaconda3 or Miniconda3 installed
REM   - Internet connection for downloading packages
REM
REM After running this script, you still need to:
REM   1. Download CloudComPy binaries from:
REM      https://www.simulation.openfields.fr/index.php/cloudcompy-downloads
REM   2. Extract to C:\bin\CloudComPy311 (or your preferred location)
REM
REM Usage: Run this script from Anaconda Prompt or any terminal with conda
REM ============================================================================

setlocal EnableDelayedExpansion

echo ============================================================================
echo CloudComPy311 Conda Environment Setup
echo ============================================================================
echo.

REM Check if conda is available
where conda >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Conda not found in PATH!
    echo         Please run this script from Anaconda Prompt or install Conda first.
    echo         Download from: https://docs.conda.io/en/latest/miniconda.html
    exit /b 1
)

echo [INFO] Conda found. Proceeding with setup...
echo.

REM Initialize conda for this shell
echo [STEP 1/5] Initializing conda...
call conda activate base
if %ERRORLEVEL% neq 0 (
    echo [WARNING] Could not activate base environment, continuing anyway...
)

REM Update conda
echo.
echo [STEP 2/5] Updating conda...
call conda update -y -n base -c defaults conda
if %ERRORLEVEL% neq 0 (
    echo [WARNING] Conda update failed, continuing anyway...
)

REM Check if environment already exists
echo.
echo [STEP 3/5] Creating CloudComPy311 environment...
call conda env list | findstr /C:"CloudComPy311" >nul 2>&1
if %ERRORLEVEL% equ 0 (
    echo [INFO] Environment CloudComPy311 already exists.
    set /p RECREATE="Do you want to remove and recreate it? (y/N): "
    if /i "!RECREATE!"=="y" (
        echo [INFO] Removing existing environment...
        call conda env remove -y -n CloudComPy311
    ) else (
        echo [INFO] Keeping existing environment. Skipping creation.
        goto install_packages
    )
)

REM Create the environment with Python 3.11
call conda create -y --name CloudComPy311 python=3.11
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to create conda environment!
    exit /b 1
)
echo [SUCCESS] Environment created.

:install_packages
echo.
echo [STEP 4/5] Installing required packages (this may take 10-30 minutes)...
echo            Please be patient...
echo.

call conda activate CloudComPy311
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to activate CloudComPy311 environment!
    exit /b 1
)

REM Configure conda-forge channel
call conda config --add channels conda-forge
call conda config --set channel_priority flexible

REM Install all required packages
REM Split into groups for better error handling
echo [INFO] Installing core packages...
call conda install -y ^
    "boost=1.84" ^
    "cgal=5.6" ^
    cmake ^
    "draco=1.5" ^
    "ffmpeg=6.1" ^
    "gdal=3.8" ^
    laszip ^
    "mpir=3.0" ^
    "mysql=8" ^
    numpy ^
    "opencv=4.9" ^
    "openmp=8.0" ^
    "openssl>=3.1" ^
    "pcl=1.14" ^
    "pdal=2.6" ^
    pybind11 ^
    "qhull=2020.2" ^
    "qt=5.15.8" ^
    scipy ^
    tbb ^
    tbb-devel ^
    "xerces-c=3.2"

if %ERRORLEVEL% neq 0 (
    echo [WARNING] Some core packages may have failed. Continuing...
)

echo.
echo [INFO] Installing Python tools and visualization packages...
call conda install -y ^
    jupyterlab ^
    notebook ^
    "matplotlib=3.9" ^
    "psutil=6.0" ^
    quaternion ^
    sphinx_rtd_theme ^
    spyder

if %ERRORLEVEL% neq 0 (
    echo [WARNING] Some Python packages may have failed. Continuing...
)

echo.
echo [STEP 5/5] Verifying installation...
python --version
echo.

echo ============================================================================
echo Setup Complete!
echo ============================================================================
echo.
echo Next steps:
echo.
echo 1. Download CloudComPy binaries from:
echo    https://www.simulation.openfields.fr/index.php/cloudcompy-downloads
echo.
echo 2. Extract the archive to: C:\bin\CloudComPy311
echo    (or update the path in run_cloudcompy.bat)
echo.
echo 3. Test the installation:
echo    conda activate CloudComPy311
echo    cd C:\bin\CloudComPy311
echo    envCloudComPy.bat
echo    python -c "import cloudComPy as cc; print('CloudComPy OK!')"
echo.
echo 4. Run your processing script:
echo    .\run_cloudcompy.bat
echo.
echo ============================================================================

endlocal
pause
