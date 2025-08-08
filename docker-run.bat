@echo off
setlocal enabledelayedexpansion

set IMAGE_NAME=web-clipboard-go
set TAG=latest
set PORT=5000

echo Starting Web Clipboard Go container...

echo.
echo Stopping existing container if running...
docker stop web-clipboard-go 2>nul || echo No existing container to stop

echo.
echo Removing existing container if exists...
docker rm web-clipboard-go 2>nul || echo No existing container to remove

echo.
echo Starting new container...
docker run -d ^
    --name web-clipboard-go ^
    --restart unless-stopped ^
    -p %PORT%:5000 ^
    %IMAGE_NAME%:%TAG%

if !errorlevel! equ 0 (
    echo.
    echo Container started successfully!
    echo Application is available at: http://localhost:%PORT%
    echo.
    echo To check logs: docker logs web-clipboard-go
    echo To stop: docker stop web-clipboard-go
) else (
    echo Failed to start container!
    exit /b 1
)