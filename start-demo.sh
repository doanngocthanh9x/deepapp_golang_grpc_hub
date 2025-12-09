#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Starting Demo System${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
# Set port
export PORT=8081
# Check if gRPC Hub is running
echo -e "${YELLOW}Checking gRPC Hub...${NC}"
if ! lsof -Pi :50051 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo -e "${RED}✗ gRPC Hub is not running on port 50051${NC}"
    echo -e "${YELLOW}Starting gRPC Hub...${NC}"
    cd "$(dirname "$0")"
    go run cmd/hub/main.go &
    HUB_PID=$!
    echo -e "${GREEN}✓ gRPC Hub started (PID: $HUB_PID)${NC}"
    sleep 2
else
    echo -e "${GREEN}✓ gRPC Hub is already running${NC}"
fi

# Start Python Worker
echo -e "${YELLOW}Starting Python Worker...${NC}"
cd "$(dirname "$0")/services/python-worker"
python3 worker.py &
PYTHON_PID=$!
echo -e "${GREEN}✓ Python Worker started (PID: $PYTHON_PID)${NC}"
sleep 1

# Start Web API
echo -e "${YELLOW}Starting Web API...${NC}"
cd "$(dirname "$0")"
go run services/web-api/main.go &
API_PID=$!
echo -e "${GREEN}✓ Web API started (PID: $API_PID)${NC}"
sleep 2

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All Services Started!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "gRPC Hub:     ${GREEN}localhost:50051${NC}"
echo -e "Python Worker: ${GREEN}Running${NC}"
echo -e "Web API:      ${GREEN}http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}Open your browser: ${GREEN}http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# Wait for Ctrl+C
trap "echo '' && echo 'Stopping services...' && kill $HUB_PID $PYTHON_PID $API_PID 2>/dev/null && echo 'All services stopped.' && exit 0" INT

# Keep script running
wait
