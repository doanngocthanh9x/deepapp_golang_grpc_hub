#!/bin/bash

# Test Java Maven Worker with v1 system
echo "Testing Java Maven Worker with v1..."

# Test Java compilation
echo "1. Testing Java compilation..."
curl -X POST http://localhost:8081/api/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "java_compile",
    "data": {
      "source": "public class Main { public static void main(String[] args) { System.out.println(\"Hello from Java!\"); } }",
      "className": "Main"
    }
  }' | jq .

echo -e "\n2. Testing Maven build..."
curl -X POST http://localhost:8081/api/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "maven_build",
    "data": {
      "goals": ["clean", "compile"],
      "pom": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><project xmlns=\"http://maven.apache.org/POM/4.0.0\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xsi:schemaLocation=\"http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd\"><modelVersion>4.0.0</modelVersion><groupId>com.test</groupId><artifactId>test</artifactId><version>1.0.0</version><properties><maven.compiler.source>17</maven.compiler.source><maven.compiler.target>17</maven.compiler.target></properties></project>"
    }
  }' | jq .

echo -e "\n3. Checking all registered capabilities..."
curl http://localhost:8081/api/capabilities | jq .

echo -e "\nJava Maven Worker v1 test completed!"