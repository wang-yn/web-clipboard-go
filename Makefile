.PHONY: build backend-build frontend-install frontend-build docker-build docker-run docker-stop docker-logs clean help run test compose-up compose-down

IMAGE_NAME := web-clipboard-go
TAG := latest
PORT := 5000
GO_ENTRY := ./backend/cmd/web-clipboard
FRONTEND_DIR := frontend

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build frontend and Go application"
	@echo "  backend-build - Build Go application only"
	@echo "  frontend-build - Install and build frontend"
	@echo "  run           - Run the application"
	@echo "  test          - Run tests"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run container"
	@echo "  docker-stop   - Stop and remove container"
	@echo "  docker-logs   - Show container logs"
	@echo "  compose-up    - Start with docker-compose"
	@echo "  compose-down  - Stop docker-compose services"
	@echo "  clean         - Clean up Docker resources and build artifacts"

# Build full application
build: frontend-build backend-build

# Build Go application
backend-build:
	@echo "Building Go application..."
	go build -o bin/web-clipboard-go.exe $(GO_ENTRY)
	@echo "Build completed! Binary: bin/web-clipboard-go.exe"

# Install frontend dependencies
frontend-install:
	@echo "Installing frontend dependencies..."
	pnpm --dir $(FRONTEND_DIR) install --frozen-lockfile

# Build frontend assets
frontend-build: frontend-install
	@echo "Building frontend..."
	pnpm --dir $(FRONTEND_DIR) build

# Run the application
run: build
	@echo "Running application..."
	WEB_CLIPBOARD_DATA_DIR=./data ./bin/web-clipboard-go.exe

# Run tests
test: frontend-build
	@echo "Running tests..."
	go test -v ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME):$(TAG) -f Dockerfile .
	@echo "Build completed!"

# Run Docker container
docker-run:
	@echo "Starting Docker container..."
	-docker stop $(IMAGE_NAME) 2>/dev/null || true
	-docker rm $(IMAGE_NAME) 2>/dev/null || true
	docker run -d --name $(IMAGE_NAME) --restart unless-stopped -p $(PORT):5000 -v ./data:/data $(IMAGE_NAME):$(TAG)
	@echo "Container started at http://localhost:$(PORT)"

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	-docker stop $(IMAGE_NAME)
	-docker rm $(IMAGE_NAME)

# Show container logs
docker-logs:
	docker logs -f $(IMAGE_NAME)

# Start with docker-compose
compose-up:
	@echo "Starting with docker-compose..."
	docker-compose up -d
	@echo "Services started at http://localhost:5000"

# Stop docker-compose services
compose-down:
	@echo "Stopping docker-compose services..."
	docker-compose down

# Clean up Docker resources and build artifacts
clean:
	@echo "Cleaning up Docker resources..."
	-docker-compose down
	-docker stop $(IMAGE_NAME) 2>/dev/null || true
	-docker rm $(IMAGE_NAME) 2>/dev/null || true
	-docker rmi $(IMAGE_NAME):$(TAG) 2>/dev/null || true
	docker system prune -f
	@echo "Cleaning up build artifacts..."
	-rm -rf bin/
	-rm -rf frontend/dist/
	@echo "Cleanup completed!"
