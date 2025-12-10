#!/bin/bash
# Test script for VietOCR Worker

echo "==========================================="
echo "VietOCR Worker Test Script"
echo "==========================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if container is running
echo -e "\n${YELLOW}1. Checking container status...${NC}"
if docker ps | grep -q "deepapp-hub-all-in-one"; then
    echo -e "${GREEN}✓ Container is running${NC}"
else
    echo -e "${RED}✗ Container is not running${NC}"
    echo "Start with: docker-compose -f docker-compose.all-in-one.yml up"
    exit 1
fi

# Check worker registration
echo -e "\n${YELLOW}2. Checking worker registration...${NC}"
sleep 5  # Wait for workers to register

WORKERS=$(curl -s http://localhost:8081/api/workers)
if echo "$WORKERS" | grep -q "python-vietocr-worker"; then
    echo -e "${GREEN}✓ VietOCR worker is registered${NC}"
    echo "$WORKERS" | jq '.workers[] | select(.worker_id=="python-vietocr-worker") | {worker_id, worker_type, capabilities: .capabilities | map(.name)}'
else
    echo -e "${RED}✗ VietOCR worker not found${NC}"
    echo "Workers found:"
    echo "$WORKERS" | jq '.workers[] | .worker_id'
fi

# Test OCR detect
echo -e "\n${YELLOW}3. Testing OCR detect capability...${NC}"

# Create a simple test image (1x1 white pixel, base64)
# In real test, you should use actual Vietnamese text image
TEST_IMAGE="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="

RESPONSE=$(curl -s -X POST http://localhost:8081/api/call \
  -H 'Content-Type: application/json' \
  -d "{
    \"worker_id\": \"python-vietocr-worker\",
    \"capability\": \"ocr_detect\",
    \"data\": {
      \"image\": \"$TEST_IMAGE\"
    }
  }")

if echo "$RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ OCR detect test passed${NC}"
    echo "Response:"
    echo "$RESPONSE" | jq '.data'
else
    echo -e "${RED}✗ OCR detect test failed${NC}"
    echo "Response:"
    echo "$RESPONSE" | jq '.'
fi

# Test OCR batch
echo -e "\n${YELLOW}4. Testing OCR batch capability...${NC}"

RESPONSE=$(curl -s -X POST http://localhost:8081/api/call \
  -H 'Content-Type: application/json' \
  -d "{
    \"worker_id\": \"python-vietocr-worker\",
    \"capability\": \"ocr_batch\",
    \"data\": {
      \"images\": [\"$TEST_IMAGE\", \"$TEST_IMAGE\"]
    }
  }")

if echo "$RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ OCR batch test passed${NC}"
    echo "Response:"
    echo "$RESPONSE" | jq '.data | {total_images, successful, total_processing_time_ms}'
else
    echo -e "${RED}✗ OCR batch test failed${NC}"
    echo "Response:"
    echo "$RESPONSE" | jq '.'
fi

# Check worker logs
echo -e "\n${YELLOW}5. Recent worker logs...${NC}"
docker logs deepapp-hub-all-in-one 2>&1 | grep -A 5 "python-vietocr-worker" | tail -20

echo -e "\n${GREEN}==========================================="
echo "Test Complete!"
echo "===========================================${NC}"
echo ""
echo "📝 Notes:"
echo "  - Worker is running in DEMO MODE (no ONNX models loaded)"
echo "  - To use real OCR, convert VietOCR models to ONNX"
echo "  - See: services/python-vietocr-worker/README.md"
echo ""
echo "🔗 Useful commands:"
echo "  View all workers: curl http://localhost:8081/api/workers | jq"
echo "  View logs: docker logs -f deepapp-hub-all-in-one"
echo "  Test with image: curl -X POST http://localhost:8081/api/call \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"worker_id\":\"python-vietocr-worker\",\"capability\":\"ocr_detect\",\"data\":{\"image\":\"BASE64_IMAGE\"}}'"
