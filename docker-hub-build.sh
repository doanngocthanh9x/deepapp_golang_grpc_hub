#!/bin/bash
# Build and push Docker image to Docker Hub

# Configuration
DOCKER_USERNAME="${DOCKER_USERNAME:-your-dockerhub-username}"
IMAGE_NAME="${IMAGE_NAME:-deepapp-hub}"
VERSION="${VERSION:-latest}"
FULL_IMAGE="${DOCKER_USERNAME}/${IMAGE_NAME}:${VERSION}"

echo "🐳 Building Docker image for Docker Hub"
echo "================================================"
echo "Image: ${FULL_IMAGE}"
echo "================================================"

# Check if logged in to Docker Hub
if ! docker info | grep -q "Username"; then
    echo "⚠️  Not logged in to Docker Hub"
    echo "Please run: docker login"
    exit 1
fi

# Build with cache busting
export CACHEBUST=$(date +%s)
echo "📦 Building with CACHEBUST=${CACHEBUST}..."

docker build \
    -f Dockerfile.all-in-one \
    -t ${FULL_IMAGE} \
    --build-arg CACHEBUST=${CACHEBUST} \
    .

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful!"
echo ""

# Ask before pushing
read -p "Push to Docker Hub? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📤 Pushing to Docker Hub..."
    docker push ${FULL_IMAGE}
    
    if [ $? -eq 0 ]; then
        echo "✅ Successfully pushed to Docker Hub!"
        echo ""
        echo "📋 Portainer Stack Configuration:"
        echo "================================================"
        echo "Image: ${FULL_IMAGE}"
        echo "Ports: 8081:8081"
        echo "Volumes: hub-data:/data"
        echo "================================================"
        echo ""
        echo "🐳 To run in Portainer:"
        echo "1. Add new stack"
        echo "2. Use docker-compose.portainer.yml"
        echo "3. Set environment variable:"
        echo "   DOCKER_IMAGE=${FULL_IMAGE}"
    else
        echo "❌ Push failed"
        exit 1
    fi
else
    echo "Skipped push"
fi
