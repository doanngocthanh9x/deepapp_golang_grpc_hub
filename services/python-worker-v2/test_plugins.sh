#!/bin/bash

echo "🧪 Testing Plugin Worker System"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test plugin loading
echo "📦 Test 1: Plugin Loading"
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub/services/python-worker-v2

python3 << 'EOF'
import sys
from plugin_loader import PluginLoader

try:
    loader = PluginLoader("plugins")
    caps = loader.load_plugins()
    print(f"✅ Loaded {len(caps)} capabilities:")
    for name in caps.keys():
        print(f"  - {name}")
except Exception as e:
    print(f"❌ Error: {e}")
    import traceback
    traceback.print_exc()
EOF

echo ""
echo "📝 Test 2: Decorator Registry"
python3 << 'EOF'
from decorators import get_registered_capabilities

caps = get_registered_capabilities()
print(f"✅ Registry has {len(caps)} capabilities:")
for name, info in caps.items():
    print(f"  - {name}: {info['description']}")
EOF

echo ""
echo "✅ Plugin system tests complete!"
