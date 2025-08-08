.PHONY: build build-minimal docker-build docker-run docker-stop docker-logs clean help

IMAGE_NAME := web-clipboard-go
TAG := latest
PORT := 5000

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build Go application"
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-run    - Run container"
	@echo "  docker-stop   - Stop and remove container"
	@echo "  docker-logs   - Show container logs"
	@echo "  compose-up    - Start with docker-compose"
	@echo "  compose-down  - Stop docker-compose services"
	@echo "  compose-nginx - Start with nginx proxy"
	@echo "  clean         - Clean up Docker resources"

# Build Go application
build:
	@echo "Building Go application..."
	go build -o web-clipboard-go.exe

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker build -t $(IMAGE_NAME):$(TAG) -f Dockerfile .
	docker build -t $(IMAGE_NAME):minimal -f Dockerfile.minimal .
	@echo "Build completed!"

# Build minimal image only
build-minimal:
	@echo "Building minimal Docker image..."
	docker build -t $(IMAGE_NAME):minimal -f Dockerfile.minimal .

# Run Docker container
docker-run:
	@echo "Starting Docker container..."
	-docker stop $(IMAGE_NAME) 2>/dev/null || true
	-docker rm $(IMAGE_NAME) 2>/dev/null || true
	docker run -d --name $(IMAGE_NAME) --restart unless-stopped -p $(PORT):5000 $(IMAGE_NAME):$(TAG)
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

# Start with nginx proxy
compose-nginx:
	@echo "Starting with nginx proxy..."
	docker-compose -f docker-compose.nginx.yml up -d
	@echo "Services started at http://localhost:8080"

# Clean up Docker resources
clean:
	@echo "Cleaning up Docker resources..."
	-docker-compose down
	-docker-compose -f docker-compose.nginx.yml down
	-docker stop $(IMAGE_NAME) 2>/dev/null || true
	-docker rm $(IMAGE_NAME) 2>/dev/null || true
	-docker rmi $(IMAGE_NAME):$(TAG) 2>/dev/null || true
	-docker rmi $(IMAGE_NAME):minimal 2>/dev/null || true
	docker system prune -f