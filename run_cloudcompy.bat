@echo off
REM Wrapper script to run process_las_files.py with CloudComPy environment
REM Usage: run_cloudcompy.bat [arguments for process_las_files.py]

setlocal EnableDelayedExpansion

REM Set CloudComPy installation path
set CLOUDCOMPY_PATH=C:\bin\CloudComPy311

REM Store the current directory BEFORE any changes
set "ORIGINAL_DIR=%cd%"
set "PYTHON_SCRIPT=%~dp0process_las_files.py"

REM Activate the CloudComPy311 conda environment
call conda activate CloudComPy311

REM Change to CloudComPy directory and set up environment
cd /d "%CLOUDCOMPY_PATH%"
call envCloudComPy.bat

REM Return to original directory
cd /d "%ORIGINAL_DIR%"

REM Run the Python script with all passed arguments
python "%PYTHON_SCRIPT%" %*

endlocal
