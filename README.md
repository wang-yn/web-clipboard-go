# Web Clipboard Go

A Go implementation of the web clipboard application, ported from the original C# version. This allows you to share text and files temporarily via short URLs.

## Features

- Share text content with auto-generated short IDs (4-6 characters)
- Upload and share files (max 50MB)
- Automatic cleanup of expired content (10 minutes TTL)
- User authentication and session management
- User management (admin-only user creation)
- Rate limiting and security protection
- CORS support for development
- Background cleanup service

## Project Structure

```
web-clipboard-go/
├── cmd/
│   └── web-clipboard/      # Application entry point
│       └── main.go
├── internal/
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/            # Data models and interfaces
│   ├── services/          # Business logic services
│   └── utils/             # Utility functions
├── web/
│   ├── static/            # Static assets (CSS, JS)
│   └── templates/         # HTML templates
├── bin/                   # Build output (gitignored)
├── data/                  # User data storage (gitignored)
└── Makefile              # Build automation
```

## Building and Running

### Using Scripts (Windows)

#### Build
```bash
build.bat
```

#### Run
```bash
run.bat
```

### Using Make

```bash
make build    # Build the application
make run      # Build and run
make test     # Run tests
```

### Manual Build

```bash
go build -o bin/web-clipboard-go.exe ./cmd/web-clipboard
```

### Manual Run

```bash
./bin/web-clipboard-go.exe
```

The server will start on port 5000: http://localhost:5000

## API Endpoints

### Authentication (Public)
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Get current user info

### Clipboard Operations (Authenticated)
- `POST /api/text` - Save text content
- `GET /api/text/{id}` - Retrieve text content
- `POST /api/file` - Upload file
- `GET /api/file/{id}` - Download file
- `DELETE /api/{id}` - Delete content
- `GET /api/cleanup` - Manual cleanup expired items

### User Management (Admin Only)
- `POST /api/users` - Create new user
- `GET /api/users` - List all users
- `GET /api/users/{id}` - Get user details
- `PUT /api/users/{id}` - Update user
- `DELETE /api/users/{id}` - Delete user
- `PUT /api/users/{id}/password` - Change user password

### Default Credentials
- Username: `admin`
- Password: `admin123`

**Important:** Change the default admin password after first login!

## Security Features

- **Authentication & Authorization**
  - Session-based authentication with bearer tokens
  - bcrypt password hashing
  - 2-hour session expiry (7 days with "remember me")
  - Admin-only user creation (no self-registration)
  - Role-based access control (admin/user)

- **File & Content Protection**
  - File type validation (blocks executable files)
  - Content scanning for suspicious patterns
  - File size limits (max 50MB)

- **Network Security**
  - Rate limiting per IP and endpoint
  - IP blocking after excessive failed attempts
  - Security headers (CSP, X-Frame-Options, etc.)
  - CORS configuration

- **Data Security**
  - Input sanitization
  - Automatic cleanup of expired content
  - Session cleanup for expired tokens

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

- `web-clipboard-go:latest` - Alpine-based image with non-root user

### Makefile Commands

```bash
make build           # Build Go application
make run             # Build and run the application
make test            # Run tests
make docker-build    # Build Docker image
make docker-run      # Run container
make docker-stop     # Stop and remove container
make docker-logs     # Show container logs
make compose-up      # Start with docker-compose
make compose-down    # Stop docker-compose services
make compose-nginx   # Start with nginx proxy
make clean           # Clean up Docker resources and build artifacts
make help            # Show all available targets
```

### Environment Variables

- `GIN_MODE=release` - Set Gin to release mode
- `PORT=5000` - Server port (default: 5000)

## Dependencies

- [Gin Web Framework](https://github.com/gin-gonic/gin) - HTTP web framework