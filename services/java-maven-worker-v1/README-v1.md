# Java Maven Worker cho v1

Phiên bản Java Maven Worker tương thích với hệ thống gRPC Hub v1.

## Khác biệt với v2

- **API endpoint**: `/api/invoke` thay vì `/api/v2/execute`
- **Capabilities endpoint**: `/api/capabilities` thay vì `/api/v2/capabilities`
- **Port**: Web API chạy trên port 8081 thay vì 8082
- **Hub port**: 50051 thay vì 50052

## Chạy với v1

```bash
# Chạy toàn bộ hệ thống v1 với Java worker
docker-compose up -d --build

# Hoặc chỉ build Java worker
make build-java-worker

# Test capabilities
chmod +x test-maven-worker-v1.sh
./test-maven-worker-v1.sh
```

## Capabilities

Giống với v2:

- `java_compile`: Compile Java code
- `maven_build`: Execute Maven commands
- `jar_analyze`: Analyze JAR files
- `java_test`: Run unit tests

## API Usage

```bash
# Invoke Java compilation
curl -X POST http://localhost:8081/api/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "java_compile",
    "data": {
      "source": "public class Main { public static void main(String[] args) { System.out.println(\"Hello!\"); } }",
      "className": "Main"
    }
  }'
```
