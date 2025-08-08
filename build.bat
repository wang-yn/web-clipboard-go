@echo off
echo Building web-clipboard-go...
go build -o web-clipboard-go.exe
if %errorlevel% == 0 (
    echo Build successful! Run with: web-clipboard-go.exe
) else (
    echo Build failed!
    exit /b 1
)