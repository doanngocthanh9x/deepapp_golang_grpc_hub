#!/bin/bash

# Build and Test Java Maven Worker
echo "Building Java Maven Worker..."

# Build the Maven project
cd services/java-maven-worker
mvn clean compile

if [ $? -eq 0 ]; then
    echo "✅ Maven compilation successful"

    # Build Docker image
    cd ../..
    docker build -f Dockerfile.worker-maven -t java-maven-worker:test .

    if [ $? -eq 0 ]; then
        echo "✅ Docker build successful"
        echo ""
        echo "To run the Maven worker:"
        echo "docker-compose -f docker-compose-v2.yml up java-maven-worker"
        echo ""
        echo "To test capabilities:"
        echo "chmod +x test-maven-worker.sh && ./test-maven-worker.sh"
    else
        echo "❌ Docker build failed"
        exit 1
    fi
else
    echo "❌ Maven compilation failed"
    exit 1
fi