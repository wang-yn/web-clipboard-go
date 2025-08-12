# Web Clipboard Go - Project Context

## Overview

This project is a Go implementation of a web-based clipboard service. It allows users to share text snippets and files temporarily via short, auto-generated URLs. Content is automatically deleted after a 10-minute time-to-live (TTL).

### Key Features

*   **Text Sharing**: Save and retrieve text content using short IDs.
*   **File Sharing**: Upload and download files (max 50MB) with short IDs.
*   **Auto-Generated IDs**: IDs are typically 4-6 characters long.
*   **Automatic Cleanup**: A background service runs every minute to remove expired content.
*   **Security Measures**:
    *   File type validation to block common executable types.
    *   Content scanning for suspicious patterns.
    *   Rate limiting per IP address and endpoint.
    *   IP blocking after excessive failed attempts.
    *   Security headers for web responses.
    *   Input sanitization.
*   **API Endpoints**: RESTful API for all core functions.
*   **Web Interface**: Serves a static HTML/JS frontend.
*   **Docker Support**: Includes configurations for standard, minimal (scratch-based), and distroless Docker images, as well as docker-compose setups (standalone, with Nginx).

### Core Technologies

*   **Language**: Go (Golang) 1.24.4
*   **Web Framework**: Gin
*   **Dependencies**: Gin framework, mimetype detection library (indirect).
*   **Deployment**: Docker, docker-compose.

## Project Structure

*   `main.go`: Application entry point, sets up the Gin router, HTTP server, and starts the cleanup service.
*   `handlers.go`: Contains the logic for all API endpoints (`saveText`, `getText`, `saveFile`, `getFile`, `deleteItem`, `cleanup`).
*   `middleware.go`: Implements CORS, security headers, and rate limiting middleware.
*   `security.go`: Houses the `SecurityService` (file/content validation, IP blocking) and `RateLimitService`.
*   `types.go`: Defines application structs like `App`, `ClipboardItem`, request/response types.
*   `utils.go`: Utility functions for ID generation and temporary directory management.
*   `wwwroot/`: Directory containing the static web frontend files (HTML, JS, CSS).
*   `nginx/`: Configuration for Nginx reverse proxy (used with docker-compose).
*   `Dockerfile*`: Various Docker image definitions.
*   `docker-compose*.yml`: Docker Compose configurations for different deployment scenarios.
*   `Makefile`: Defines common build, run, and Docker commands.
*   `go.mod`, `go.sum`: Go module dependency files.

## Building and Running

### Local Development

1.  **Build**: `go build -o web-clipboard-go.exe` (or `make build`)
2.  **Run**: `./web-clipboard-go.exe`
    *   The server starts on port 5000: `http://localhost:5000`.

### Docker

1.  **Build Images**: `make docker-build` (builds standard, minimal, and distroless images).
2.  **Run Container**: `make docker-run` (uses the standard image).
3.  **Run with Docker Compose**:
    *   Standard: `docker-compose up -d`
    *   With Nginx: `docker-compose -f docker-compose.nginx.yml up -d`
    *   Production: `docker-compose -f docker-compose.prod.yml up -d`
4.  **Makefile Commands**: The `Makefile` provides convenient shortcuts for common Docker tasks.

### Environment Variables

*   `GIN_MODE=release`: Sets Gin to release mode (recommended for production).
*   `PORT=5000`: Defines the server port (default is 5000).

## API Endpoints

*   `POST /api/text`: Save text content. Request body: `{ "content": "..." }`. Response: `{ "id": "...", "expiresAt": "..." }`.
*   `GET /api/text/{id}`: Retrieve text content by ID. Response: `{ "content": "...", "createdAt": "..." }`.
*   `POST /api/file`: Upload a file (multipart form data, field name 'file'). Response: `{ "id": "...", "fileName": "...", "expiresAt": "..." }`.
*   `GET /api/file/{id}`: Download a file by ID.
*   `DELETE /api/{id}`: Delete any content (text or file) by ID.
*   `GET /api/cleanup`: Manually trigger the expired item cleanup process.

## Development Notes

*   **Concurrency**: Uses `sync.RWMutex` to protect the in-memory `clipboardData` map.
*   **Temporary Files**: Uploaded files are stored in a temporary directory (`os.TempDir() + "/web-clipboard-go"`).
*   **Configuration**: Most configuration is hardcoded (e.g., TTL, rate limits, blocked extensions), making it easy to understand but less flexible without code changes.