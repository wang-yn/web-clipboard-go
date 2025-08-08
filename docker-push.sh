#!/bin/bash

# Check if a version is provided
if [ -z "$1" ]; then
  echo "Error: No version specified."
  echo "Usage: ./docker-push.sh <version>"
  exit 1
fi

# Assign the version from the first argument
ver=$1

echo "Pushing version: $ver"

# Tag images with the new version
docker tag web-clipboard-go:minimal "yiwayhb/web-clipboard-go:$ver"
docker tag web-clipboard-go:minimal yiwayhb/web-clipboard-go:minimal
docker tag web-clipboard-go:minimal yiwayhb/web-clipboard-go:latest

# Push images to the registry
docker push "yiwayhb/web-clipboard-go:$ver"
docker push yiwayhb/web-clipboard-go:minimal
docker push yiwayhb/web-clipboard-go:latest

echo "Push complete for version $ver, minimal, and latest."
