#!/bin/bash
# Test script để verify trước khi push lên GitHub

set -e

echo "🧪 Testing Docker Build & Deploy"
echo "================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check Docker
echo -e "\n${YELLOW}[1/6]${NC} Checking Docker..."
if ! command -v docker &> /dev/null; then
    echo -e "${RED}✗ Docker not found${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker installed${NC}"

# Test 2: Check docker-compose
echo -e "\n${YELLOW}[2/6]${NC} Checking docker-compose..."
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}✗ docker-compose not found${NC}"
    exit 1
fi
echo -e "${GREEN}✓ docker-compose installed${NC}"

# Test 3: Build images
echo -e "\n${YELLOW}[3/6]${NC} Building Docker images..."
docker-compose build
echo -e "${GREEN}✓ Images built successfully${NC}"

# Test 4: Start services
echo -e "\n${YELLOW}[4/6]${NC} Starting services..."
docker-compose down -v 2>/dev/null || true
docker-compose up -d
echo -e "${GREEN}✓ Services started${NC}"

# Wait for services
echo -e "\n${YELLOW}Waiting for services to be ready...${NC}"
sleep 10

# Test 5: Test endpoints
echo -e "\n${YELLOW}[5/6]${NC} Testing API endpoints..."

# Test status
echo -n "  Testing /api/status... "
if curl -sf http://localhost:8081/api/status > /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    docker-compose logs webapi
    exit 1
fi

# Test hello
echo -n "  Testing /api/hello... "
if curl -sf -X POST http://localhost:8081/api/hello > /dev/null; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    docker-compose logs webapi worker
    exit 1
fi

echo -e "${GREEN}✓ All endpoints working${NC}"

# Test 6: Check logs
echo -e "\n${YELLOW}[6/6]${NC} Checking service logs..."
echo ""
echo "Hub logs:"
docker-compose logs --tail=5 hub
echo ""
echo "Worker logs:"
docker-compose logs --tail=5 worker
echo ""
echo "Web API logs:"
docker-compose logs --tail=5 webapi

# Cleanup
echo -e "\n${YELLOW}Cleaning up...${NC}"
docker-compose down

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}✓ All tests passed!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "Ready to push to GitHub:"
echo "  git add ."
echo "  git commit -m 'Add Docker support and CI/CD'"
echo "  git push origin main"
echo ""
echo "After push, setup Docker Hub secrets in GitHub:"
echo "  DOCKER_HUB_USERNAME"
echo "  DOCKER_HUB_TOKEN"
