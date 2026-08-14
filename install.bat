@echo off
echo ========================================
echo   FastSS Global Installer for Windows
echo ========================================

echo 1. Compiling binary...
go build -o fastss.exe .
if errorlevel 1 (
    echo [ERROR] Build failed! Make sure Go is installed.
    exit /b 1
)

copy /Y fastss.exe ss.exe >nul

echo 2. Installing to Global PATH (%USERPROFILE%\go\bin)...
if not exist "%USERPROFILE%\go\bin" (
    mkdir "%USERPROFILE%\go\bin"
)

copy /Y ss.exe "%USERPROFILE%\go\bin\ss.exe" >nul
copy /Y fastss.exe "%USERPROFILE%\go\bin\fastss.exe" >nul

echo.
echo =======================================================
echo [SUCCESS] FastSS installed successfully!
echo You can now use the 'ss' command from ANY folder in CMD or PowerShell.
echo (If currently open, restart your CMD/PowerShell window to refresh PATH).
echo =======================================================
