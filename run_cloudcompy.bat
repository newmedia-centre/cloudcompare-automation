@echo off
REM Wrapper script to run process_las_files.py with CloudComPy environment
REM Usage: run_cloudcompy.bat [arguments for process_las_files.py]

setlocal EnableDelayedExpansion

REM Set CloudComPy installation path
set CLOUDCOMPY_PATH=C:\bin\CloudComPy311

REM Store the current directory BEFORE any changes
set "ORIGINAL_DIR=%cd%"
set "PYTHON_SCRIPT=%~dp0process_las_files.py"

REM Check if conda is available
where conda >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Conda not found in PATH
    echo [ERROR] Please run from Anaconda Prompt or add conda to PATH
    exit /b 1
)

REM Activate the CloudComPy311 conda environment
call conda activate CloudComPy311 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to activate CloudComPy311 conda environment
    echo [ERROR] Run setup_cloudcompy.bat first to create the environment
    exit /b 1
)

REM Check if CloudComPy path exists
if not exist "%CLOUDCOMPY_PATH%\envCloudComPy.bat" (
    echo [ERROR] CloudComPy not found at %CLOUDCOMPY_PATH%
    echo [ERROR] Download CloudComPy from https://www.simulation.openfields.fr/index.php/cloudcompy-downloads
    echo [ERROR] Extract to %CLOUDCOMPY_PATH%
    exit /b 1
)

REM Change to CloudComPy directory and set up environment
cd /d "%CLOUDCOMPY_PATH%"
call envCloudComPy.bat >nul 2>&1

REM Quick test that cloudComPy can be imported
echo [INFO] Checking environment, Python test: import cloudComPy
python -c "import cloudComPy" 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] CloudComPy import failed
    echo [ERROR] The CloudComPy environment may not be set up correctly
    cd /d "%ORIGINAL_DIR%"
    exit /b 1
)
echo [INFO] Environment OK!

REM Return to original directory
cd /d "%ORIGINAL_DIR%"

REM Run the Python script with all passed arguments
python "%PYTHON_SCRIPT%" %*

endlocal
