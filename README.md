# Web Clipboard Go

A Go implementation of the web clipboard application, ported from the original C# version. This allows you to share text and files temporarily via short URLs.

## Features

- Share text content with auto-generated short IDs (4-6 characters)
- Upload and share files (max 50MB)
- Automatic cleanup of expired content (10 minutes TTL)
- Rate limiting and security protection
- CORS support for development
- Background cleanup service

## Building and Running

### Build
```bash
go build -o web-clipboard-go.exe
```

### Run
```bash
./web-clipboard-go.exe
```

The server will start on port 5000: http://localhost:5000

## API Endpoints

- `POST /api/text` - Save text content
- `GET /api/text/{id}` - Retrieve text content
- `POST /api/file` - Upload file
- `GET /api/file/{id}` - Download file
- `DELETE /api/{id}` - Delete content
- `GET /api/cleanup` - Manual cleanup expired items

## Security Features

- File type validation (blocks executable files)
- Content scanning for suspicious patterns
- Rate limiting per IP
- IP blocking after failed attempts
- Security headers
- Input sanitization

## Docker Support

### Quick Start with Docker

Build and run with Docker:
```bash
# Build Docker images
./docker-build.bat  # Windows
./docker-build.sh   # Linux/Mac

# Run container
docker run -p 5000:5000 web-clipboard-go:latest
```

### Docker Compose

Start with docker-compose:
```bash
# Standard deployment
docker-compose up -d

# With Nginx reverse proxy
docker-compose -f docker-compose.nginx.yml up -d

# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

### Available Docker Images

- `web-clipboard-go:latest` - Standard Alpine-based image (~35MB)
- `web-clipboard-go:minimal` - Minimal scratch-based image (~13MB)  
- `web-clipboard-go:distroless` - Google distroless image (~14MB)

### Makefile Commands

```bash
make docker-build     # Build Docker images
make docker-run       # Run container
make compose-up       # Start with docker-compose
make compose-nginx    # Start with nginx proxy
make clean           # Clean up Docker resources
```

### Environment Variables

- `GIN_MODE=release` - Set Gin to release mode
- `PORT=5000` - Server port (default: 5000)

## Dependencies

- [Gin Web Framework](https://github.com/gin-gonic/gin) - HTTP web framework