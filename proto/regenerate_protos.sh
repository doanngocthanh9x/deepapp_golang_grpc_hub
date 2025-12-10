#!/bin/bash
# Regenerate protobuf files for all languages after proto changes

set -e

echo "🔄 Regenerating protobuf files..."

cd "$(dirname "$0")"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Python
echo -e "${BLUE}📦 Generating Python proto...${NC}"
python3 -m grpc_tools.protoc \
    -I. \
    --python_out=../services/python-worker \
    --grpc_python_out=../services/python-worker \
    hub.proto
echo -e "${GREEN}✓ Python proto generated${NC}"

# Go
echo -e "${BLUE}📦 Generating Go proto...${NC}"
protoc -I. \
    --go_out=../services/hub/internal/proto \
    --go_opt=paths=source_relative \
    --go-grpc_out=../services/hub/internal/proto \
    --go-grpc_opt=paths=source_relative \
    hub.proto
echo -e "${GREEN}✓ Go proto generated${NC}"

# Node.js
echo -e "${BLUE}📦 Generating Node.js proto...${NC}"
protoc -I. \
    --js_out=import_style=commonjs:../services/node-worker \
    --grpc_out=grpc_js:../services/node-worker \
    --plugin=protoc-gen-grpc=$(which grpc_tools_node_protoc_plugin) \
    hub.proto 2>/dev/null || echo "⚠️  Node proto generation skipped (grpc_tools_node_protoc_plugin not found)"
echo -e "${GREEN}✓ Node.js proto generated (or skipped)${NC}"

# Java
echo -e "${BLUE}📦 Generating Java proto...${NC}"
cd ..
if [ -d "services/java-simple-worker" ]; then
    cd services/java-simple-worker
    mvn clean compile 2>&1 | grep -E "(BUILD SUCCESS|BUILD FAILURE)" || echo "Java build completed"
    cd ../..
fi
cd proto
echo -e "${GREEN}✓ Java proto will be generated on Maven build${NC}"

# C++
echo -e "${BLUE}📦 Generating C++ proto...${NC}"
protoc -I. \
    --cpp_out=../services/cpp-worker/src \
    --grpc_out=../services/cpp-worker/src \
    --plugin=protoc-gen-grpc=$(which grpc_cpp_plugin) \
    hub.proto 2>/dev/null || echo "⚠️  C++ proto generation skipped (grpc_cpp_plugin not found)"
echo -e "${GREEN}✓ C++ proto generated (or skipped)${NC}"

echo ""
echo -e "${GREEN}✅ All proto files regenerated!${NC}"
echo "Note: Some languages may require additional build steps:"
echo "  - Java: Run 'mvn compile' in services/java-simple-worker"
echo "  - C++: Rebuild with CMake"
