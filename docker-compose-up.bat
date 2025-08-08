@echo off
echo Starting Web Clipboard Go with Docker Compose...

set COMPOSE_FILE=docker-compose.yml

if "%1"=="nginx" (
    set COMPOSE_FILE=docker-compose.nginx.yml
    echo Using Nginx reverse proxy configuration...
) else (
    echo Using standard configuration...
)

echo.
docker-compose -f %COMPOSE_FILE% up -d

if %errorlevel% equ 0 (
    echo.
    echo Services started successfully!
    
    if "%1"=="nginx" (
        echo Application is available at:
        echo   - HTTP: http://localhost:8080
        echo   - HTTPS: https://localhost:8443 ^(if SSL configured^)
    ) else (
        echo Application is available at: http://localhost:5000
    )
    
    echo.
    echo To check status: docker-compose -f %COMPOSE_FILE% ps
    echo To view logs: docker-compose -f %COMPOSE_FILE% logs -f
    echo To stop: docker-compose -f %COMPOSE_FILE% down
) else (
    echo Failed to start services!
    exit /b 1
)