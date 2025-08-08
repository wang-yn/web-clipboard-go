#!/bin/bash

set -e

echo "Building Docker images for Web Clipboard Go..."

IMAGE_NAME="web-clipboard-go"
TAG="latest"

echo ""
echo "Building standard image..."
docker build -t "${IMAGE_NAME}:${TAG}" -f Dockerfile .

echo ""
echo "Building minimal image..."
docker build -t "${IMAGE_NAME}:minimal" -f Dockerfile.minimal .

echo ""
echo "Building distroless image..."
docker build -t "${IMAGE_NAME}:distroless" -f Dockerfile.distroless .

echo ""
echo "Build completed successfully!"
echo ""
echo "Available images:"
docker images "${IMAGE_NAME}"

echo ""
echo "To run the application:"
echo "  docker run -p 5000:5000 ${IMAGE_NAME}:${TAG}"
echo "  docker run -p 5000:5000 ${IMAGE_NAME}:minimal"
echo "  docker run -p 5000:5000 ${IMAGE_NAME}:distroless"
echo ""
echo "Or use docker-compose:"
echo "  docker-compose up -d"
echo "  docker-compose -f docker-compose.nginx.yml up -d"