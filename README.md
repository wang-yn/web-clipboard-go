# Web Clipboard Go

A Go implementation of a temporary web clipboard for authenticated text and file sharing. Items expire automatically after 10 minutes.

## Features

- Save text and files behind authenticated API endpoints
- React-based frontend served as static assets by the Go server
- Recent item actions for copying text and downloading files
- User authentication, password changes, and admin user management
- File validation, content checks, rate limiting, and security headers
- Docker and Docker Compose deployment support

## Project Structure

```
web-clipboard-go/
├── backend/
│   ├── cmd/web-clipboard/     # Go application entry point
│   └── internal/              # handlers, middleware, models, services, utils
├── frontend/
│   ├── static/                # React UMD bundles, JSX entry files, favicon
│   └── templates/             # HTML page shells
├── .github/workflows/         # GHCR image build workflow
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── web_contract_test.go       # Static frontend/router contract tests
```

Generated binaries and runtime data belong in ignored paths such as `bin/` and `data/`.

## Building and Running

Use Make for local development:

```bash
make build    # Build bin/web-clipboard-go.exe from ./backend/cmd/web-clipboard
make run      # Build and run on http://localhost:5000
make test     # Run go test -v ./...
```

Manual commands:

```bash
go test ./...
go build -o bin/web-clipboard-go.exe ./backend/cmd/web-clipboard
./bin/web-clipboard-go.exe
```

The default admin account is created on first startup:

- Username: `admin`
- Password: `admin123`

Change this password after first login.

## API Endpoints

Authentication:

- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`

Clipboard:

- `POST /api/text`
- `GET /api/text/{id}`
- `POST /api/file`
- `GET /api/file/{id}`
- `DELETE /api/{id}`
- `GET /api/cleanup`

User management:

- `POST /api/users`
- `GET /api/users`
- `GET /api/users/{id}`
- `PUT /api/users/{id}`
- `DELETE /api/users/{id}`
- `PUT /api/users/{id}/password`

## Docker

Pull and run the published GHCR image:

```bash
docker pull ghcr.io/wang-yn/web-clipboard-go:latest
docker run -d \
  --name web-clipboard-go \
  --restart unless-stopped \
  -p 5000:5000 \
  -e GIN_MODE=release \
  ghcr.io/wang-yn/web-clipboard-go:latest
```

Local image development:

```bash
make docker-build
make docker-run
```

Docker Compose:

```bash
docker compose up -d
docker compose logs -f
docker compose down
```

## Makefile Commands

```bash
make build          # Build Go application
make run            # Build and run locally
make test           # Run tests
make docker-build   # Build Docker image
make docker-run     # Run container
make docker-stop    # Stop and remove container
make docker-logs    # Show container logs
make compose-up     # Start docker-compose services
make compose-down   # Stop docker-compose services
make clean          # Clean Docker resources and build artifacts
make help           # Show available targets
```

## Security Notes

- Do not commit secrets, generated binaries, or runtime `data/`.
- Preserve file validation, content scanning, rate limits, session expiry, and last-admin protection when changing auth or user-management code.
- Reverse proxies must forward `/`, `/login.html`, `/api/*`, `/static/*`, and `/favicon.ico`.
