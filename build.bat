@echo off
echo Building web-clipboard-go...
go build -o bin\web-clipboard-go.exe .\cmd\web-clipboard
if %errorlevel% == 0 (
    echo Build successful! Binary: bin\web-clipboard-go.exe
    echo Run with: bin\web-clipboard-go.exe
) else (
    echo Build failed!
    exit /b 1
)