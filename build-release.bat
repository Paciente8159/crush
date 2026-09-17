@echo off
setlocal enabledelayedexpansion

set CGO_ENABLED=0
set GOEXPERIMENT=greenteagc
set GOOS=windows
set GOARCH=amd64

for /f %%i in ('git describe --long 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=devel

set LDFLAGS=-s -w -X github.com/charmbracelet/crush/internal/version.Version=%VERSION%

echo Building Crush %VERSION% for windows/amd64...
go build -ldflags="%LDFLAGS%" -trimpath -o release\crush.exe .

if %ERRORLEVEL% equ 0 (
    echo Done: release\crush.exe
) else (
    echo Build failed.
    exit /b 1
)