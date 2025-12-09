# Java Maven Worker

Java-based worker for the gRPC Hub system, providing Maven and Java-related capabilities.

## Capabilities

### 1. java_compile

Compile Java source code using the Java Compiler API.

**Input:**

```json
{
  "source": "public class Main { public static void main(String[] args) { System.out.println(\"Hello!\"); } }",
  "className": "Main"
}
```

**Output:**

```json
{
  "success": true,
  "output": "Compilation successful\nGenerated: Main.class",
  "errors": []
}
```

### 2. maven_build

Execute Maven build commands.

**Input:**

```json
{
  "goals": ["clean", "compile"],
  "pom": "<?xml version=\"1.0\" encoding=\"UTF-8\"?>..."
}
```

**Output:**

```json
{
  "success": true,
  "output": "[INFO] Building project...\n[INFO] BUILD SUCCESS",
  "exitCode": 0
}
```

### 3. jar_analyze

Analyze JAR file contents and metadata.

**Input:**

```json
{
  "jarPath": "/path/to/file.jar"
}
```

**Output:**

```json
{
  "manifest": {"Main-Class": "com.example.Main"},
  "entries": ["META-INF/", "com/example/Main.class"],
  "size": 1024000
}
```

### 4. java_test

Run Java unit tests.

**Input:**

```json
{
  "testClass": "public class CalculatorTest { @Test public void testAdd() { assert 2+2 == 4; } }",
  "className": "CalculatorTest"
}
```

**Output:**

```json
{
  "success": true,
  "testsRun": 1,
  "failures": 0,
  "output": "Running 1 tests...\n✓ Test 1 passed"
}
```

## Building and Running

### Using Docker Compose

```bash
docker-compose -f docker-compose-v2.yml up java-maven-worker
```

### Manual Build

```bash
cd services/java-maven-worker
mvn clean package
java -jar target/java-maven-worker-1.0.0.jar
```

### Environment Variables

- `HUB_ADDRESS`: gRPC Hub address (default: localhost:50051)
- `WORKER_ID`: Unique worker identifier (default: auto-generated)

## Architecture

- **MavenWorker**: Main worker class managing gRPC connection and capability routing
- **CapabilityHandler**: Interface for implementing specific capabilities
- **Handlers**: Individual capability implementations (JavaCompileHandler, MavenBuildHandler, etc.)

The worker uses bidirectional gRPC streaming to communicate with the Hub and automatically registers its capabilities on startup.
