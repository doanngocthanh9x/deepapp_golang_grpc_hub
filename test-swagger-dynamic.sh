#!/bin/bash

echo "🎉 Testing Dynamic Swagger System"
echo "=================================="
echo ""

echo "📍 Endpoints Available:"
echo "  1. Dynamic Swagger JSON: http://localhost:8082/api/v2/swagger.json"
echo "  2. Dynamic Swagger UI:   http://localhost:8082/docs"
echo "  3. Static Swagger UI:    http://localhost:8082/swagger/index.html"
echo "  4. API Status:           http://localhost:8082/api/v2/status"
echo ""

echo "🧪 Test 1: Swagger JSON from Workers"
echo "======================================"
SPEC=$(curl -s http://localhost:8082/api/v2/swagger.json)
TITLE=$(echo "$SPEC" | grep -o '"title":"[^"]*"' | head -1)
PATHS_COUNT=$(echo "$SPEC" | grep -o '"/invoke/[^"]*"' | wc -l)

echo "Swagger Title: $TITLE"
echo "Dynamic Endpoints: $PATHS_COUNT capabilities"
echo ""

echo "📋 Capabilities Discovered:"
echo "$SPEC" | grep -o '"/invoke/[^"]*"' | sed 's|"/invoke/||g' | sed 's|"||g' | while read cap; do
    echo "  ✅ $cap"
done
echo ""

echo "🧪 Test 2: Check Schema Definitions"
echo "===================================="
echo "Calculator schema:"
curl -s http://localhost:8082/api/v2/swagger.json | \
    grep -A 20 '"/invoke/calculator"' | \
    grep -E '(enum|properties|required)' | head -5
echo ""

echo "🧪 Test 3: Test Via API"
echo "======================="
echo "Testing calculator (7 * 8):"
RESULT=$(curl -s -X POST http://localhost:8082/api/v2/invoke/calculator \
    -H "Content-Type: application/json" \
    -d '{"operation":"multiply","a":7,"b":8}')
echo "Response: $RESULT"
echo ""

echo "Testing text_transform (UPPERCASE):"
RESULT=$(curl -s -X POST http://localhost:8082/api/v2/invoke/text_transform \
    -H "Content-Type: application/json" \
    -d '{"text":"hello swagger","operation":"uppercase"}')
echo "Response: $RESULT"
echo ""

echo "✅ All Tests Complete!"
echo ""
echo "🌐 Open in Browser:"
echo "  Dynamic Swagger UI: http://localhost:8082/docs"
echo "  (Auto-refreshes every 30 seconds to pick up new capabilities)"
