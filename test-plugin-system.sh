#!/bin/bash

echo "🚀 Starting Full System Test with Plugin Worker"
echo "==============================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Kill old processes
echo -e "${YELLOW}🧹 Cleaning up old processes...${NC}"
pkill -9 -f "hub-v2" 2>/dev/null
pkill -9 -f "web-api-v2" 2>/dev/null
pkill -9 -f "worker_plugin" 2>/dev/null
sleep 1

# Start Hub
echo -e "${YELLOW}📡 Starting Hub on port 50052...${NC}"
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
PORT=50052 ./bin/hub-v2 > /tmp/hub_test.log 2>&1 &
HUB_PID=$!
sleep 2

if ps -p $HUB_PID > /dev/null; then
    echo -e "${GREEN}✅ Hub started (PID: $HUB_PID)${NC}"
else
    echo -e "${RED}❌ Hub failed to start${NC}"
    exit 1
fi

# Start Plugin Worker
echo -e "${YELLOW}🔌 Starting Plugin Worker...${NC}"
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub/services/python-worker-v2
HUB_ADDRESS=localhost:50052 python3 worker_plugin.py > /tmp/worker_test.log 2>&1 &
WORKER_PID=$!
sleep 3

if ps -p $WORKER_PID > /dev/null; then
    echo -e "${GREEN}✅ Worker started (PID: $WORKER_PID)${NC}"
    # Show worker output
    echo "   Worker output:"
    head -10 /tmp/worker_test.log | sed 's/^/   /'
else
    echo -e "${RED}❌ Worker failed to start${NC}"
    kill $HUB_PID 2>/dev/null
    exit 1
fi

# Start API
echo -e "${YELLOW}🌐 Starting Web API on port 8082...${NC}"
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
HUB_ADDRESS=localhost:50052 PORT=8082 ./bin/web-api-v2 > /tmp/api_test.log 2>&1 &
API_PID=$!
sleep 3

if ps -p $API_PID > /dev/null; then
    echo -e "${GREEN}✅ API started (PID: $API_PID)${NC}"
else
    echo -e "${RED}❌ API failed to start${NC}"
    kill $HUB_PID $WORKER_PID 2>/dev/null
    exit 1
fi

echo ""
echo -e "${GREEN}✅ All services started successfully!${NC}"
echo ""
echo "================================================"
echo "🧪 Testing Plugin Capabilities"
echo "================================================"
echo ""

# Test 1: Hello
echo -e "${YELLOW}Test 1: Hello Plugin${NC}"
RESULT=$(curl -s -X POST http://localhost:8082/api/v2/invoke/hello \
  -H "Content-Type: application/json" \
  -d '{"name":"Plugin System"}')
echo "Response: $RESULT"
if echo "$RESULT" | grep -q "Hello"; then
    echo -e "${GREEN}✅ PASSED${NC}"
else
    echo -e "${RED}❌ FAILED${NC}"
fi
echo ""

# Test 2: Echo
echo -e "${YELLOW}Test 2: Echo Plugin${NC}"
RESULT=$(curl -s -X POST http://localhost:8082/api/v2/invoke/echo \
  -H "Content-Type: application/json" \
  -d '{"message":"Testing echo"}')
echo "Response: $RESULT"
if echo "$RESULT" | grep -q "echo"; then
    echo -e "${GREEN}✅ PASSED${NC}"
else
    echo -e "${RED}❌ FAILED${NC}"
fi
echo ""

# Test 3: Text Transform
echo -e "${YELLOW}Test 3: Text Transform Plugin${NC}"
RESULT=$(curl -s -X POST http://localhost:8082/api/v2/invoke/text_transform \
  -H "Content-Type: application/json" \
  -d '{"text":"hello world","operation":"uppercase"}')
echo "Response: $RESULT"
if echo "$RESULT" | grep -q "HELLO WORLD"; then
    echo -e "${GREEN}✅ PASSED${NC}"
else
    echo -e "${RED}❌ FAILED${NC}"
fi
echo ""

# Test 4: Calculator
echo -e "${YELLOW}Test 4: Calculator Plugin${NC}"
RESULT=$(curl -s -X POST http://localhost:8082/api/v2/invoke/calculator \
  -H "Content-Type: application/json" \
  -d '{"operation":"multiply","a":7,"b":8}')
echo "Response: $RESULT"
if echo "$RESULT" | grep -q "56"; then
    echo -e "${GREEN}✅ PASSED${NC}"
else
    echo -e "${RED}❌ FAILED${NC}"
fi
echo ""

# Test 5: Statistics
echo -e "${YELLOW}Test 5: Statistics Plugin${NC}"
RESULT=$(curl -s -X POST http://localhost:8082/api/v2/invoke/statistics \
  -H "Content-Type: application/json" \
  -d '{"numbers":[10,20,30,40,50]}')
echo "Response: $RESULT"
if echo "$RESULT" | grep -q "mean"; then
    echo -e "${GREEN}✅ PASSED${NC}"
else
    echo -e "${RED}❌ FAILED${NC}"
fi
echo ""

echo "================================================"
echo -e "${GREEN}✅ All tests completed!${NC}"
echo "================================================"
echo ""
echo "📊 Services Status:"
echo "  - Hub: http://localhost:50052 (PID: $HUB_PID)"
echo "  - Worker: PID $WORKER_PID (7 capabilities loaded)"
echo "  - API: http://localhost:8082 (PID: $API_PID)"
echo "  - Swagger: http://localhost:8082/swagger/index.html"
echo ""
echo "📝 Logs:"
echo "  - Hub: /tmp/hub_test.log"
echo "  - Worker: /tmp/worker_test.log"
echo "  - API: /tmp/api_test.log"
echo ""
echo "🛑 To stop all services:"
echo "  kill $HUB_PID $WORKER_PID $API_PID"
