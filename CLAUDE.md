# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Web Clipboard Go is a temporary content sharing service that allows users to share text and files via short URLs. Content expires after 10 minutes. This is a Go implementation ported from an original C# version.

## Build and Run Commands

### Local Development
```bash
# Build the application
go build -o web-clipboard-go.exe

# Run the application
./web-clipboard-go.exe
# Server starts on http://localhost:5000
```

### Docker
```bash
# Build Docker images
make docker-build

# Run container
make docker-run

# View logs
make docker-logs

# Stop container
make docker-stop

# Clean up Docker resources
make clean
```

### Docker Compose
```bash
# Standard deployment
make compose-up

# With Nginx reverse proxy
make compose-nginx

# Stop services
make compose-down
```

## Architecture

### Core Components

**App struct (types.go:8)** - Central application state containing:
- `clipboardData`: In-memory map storing all clipboard items (text/files)
- `dataMutex`: RWMutex for thread-safe access to clipboard data
- `tempDir`: Directory for temporary file storage
- `security`: Security service handling validation and IP blocking
- `rateLimiter`: Rate limiting service per IP/endpoint
- `cleanupTicker`: Background ticker for expired item cleanup

### Service Architecture

**SecurityService (security.go:14)** - Handles all security concerns:
- IP-based blocking after failed attempts (>20 blocks IP)
- File type validation (blocks executables: .exe, .bat, .cmd, .dll, .php, etc.)
- Content scanning for suspicious patterns (XSS, path traversal, eval, etc.)
- Request validation with automatic blocking
- Failed attempt tracking with 1-hour expiration

**RateLimitService (security.go:179)** - Per-IP, per-endpoint rate limiting:
- POST endpoints: 20 requests/minute
- GET endpoints: 100 requests/minute
- Other endpoints: 50 requests/minute (default)
- 1-minute sliding window
- 2-minute cleanup of expired entries

### Data Flow

**Text Sharing:**
1. Client POSTs to `/api/text` with JSON content
2. `saveText` handler (handlers.go:15) validates content through SecurityService
3. Generates 4-6 character alphanumeric ID via `generateShortID` (utils.go:13)
4. Stores ClipboardItem in memory with 10-minute TTL
5. Client GETs from `/api/text/:id` to retrieve

**File Sharing:**
1. Client POSTs multipart form to `/api/file`
2. `saveFile` handler (handlers.go:73) validates file type and size (max 50MB)
3. Saves file to temp directory with format: `{id}_{filename}`
4. Stores metadata in memory with 10-minute TTL
5. Client GETs from `/api/file/:id` for download

### Middleware Chain (main.go:59)

Applied in order:
1. `corsMiddleware` - CORS for localhost origins only
2. `securityHeadersMiddleware` - Sets X-Content-Type-Options, X-Frame-Options, CSP
3. `rateLimitMiddleware` - IP-based rate limiting

### Background Services

**Cleanup Service (main.go:97)** - Runs every 1 minute:
- Deletes expired ClipboardItems from memory
- Removes associated files from disk
- Cleans expired security tracking (1-hour old)
- Cleans expired rate limit tracking (2-minute old)

**Graceful Shutdown (main.go:44)** - Listens for SIGINT/SIGTERM:
- Stops cleanup ticker
- 5-second timeout for in-flight requests
- Ensures clean application exit

### Thread Safety

All shared state uses RWMutex for concurrent access:
- `App.dataMutex` protects clipboard data map
- `SecurityService.mutex` protects failed attempts and blocked IPs
- `RateLimitService.mutex` protects rate limit tracking

### ID Generation

Short IDs are cryptographically random (utils.go:13):
- First attempts 4-character ID (100 retries)
- Falls back to 6-character if collisions occur
- Character set: lowercase a-z and 0-9 (36 chars)

### Temporary File Management

Files stored in system temp directory (utils.go:44):
- Path: `os.TempDir()/web-clipboard-go/`
- Cleaned on startup via `initTempDir` (utils.go:48)
- Deleted on item expiration or manual deletion

## Important Constraints

- **Content TTL**: Hard-coded 10 minutes for all items
- **Max File Size**: 50MB (handlers.go:92)
- **Max Text Size**: 1MB (security.go:59)
- **Port**: Hard-coded :5000 (main.go:27)
- **Request Timeouts**: 10s read, 10s write (main.go:29-30)
- **Blocked File Types**: Executables and server-side scripts (security.go:29-34)

## Security Considerations

When modifying security features:
- Never weaken file type validation - executable blocking is critical
- Rate limits prevent abuse - adjust carefully
- IP blocking is aggressive (>20 failures) - balance security vs usability
- All user content goes through suspicious pattern scanning
- Client IP extracted from X-Forwarded-For or X-Real-IP headers (security.go:129)

## Concurrency Patterns

- Always acquire locks before accessing shared maps
- Use RLock for reads, Lock for writes
- Cleanup operations must lock before iterating maps
- Handler methods are called concurrently by Gin
