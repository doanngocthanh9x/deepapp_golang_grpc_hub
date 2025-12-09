#!/bin/bash
# Rebuild script to ensure fresh Docker images without cache

echo "🔨 Starting fresh rebuild at $(date)"
echo "📝 Using CACHEBUST=$(date +%s) to force fresh build"

# Export CACHEBUST with current timestamp to force rebuild
export CACHEBUST=$(date +%s)

# Build all services
echo "🏗️  Building all services..."
docker-compose build --no-cache

# Restart services
echo "🔄 Restarting services..."
docker-compose down
docker-compose up -d

echo "✅ Rebuild complete!"
echo "📊 Check logs with: docker-compose logs -f"
