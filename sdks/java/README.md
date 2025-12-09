# DeepApp gRPC Hub - Java Worker SDK

This SDK provides a simple way to create workers that connect to the DeepApp gRPC Hub system using Java.

## Installation

Add the following dependencies to your `pom.xml`:

```xml
<dependencies>
    <!-- gRPC dependencies -->
    <dependency>
        <groupId>io.grpc</groupId>
        <artifactId>grpc-netty-shaded</artifactId>
        <version>1.60.0</version>
    </dependency>
    <dependency>
        <groupId>io.grpc</groupId>
        <artifactId>grpc-protobuf</artifactId>
        <version>1.60.0</version>
    </dependency>
    <dependency>
        <groupId>io.grpc</groupId>
        <artifactId>grpc-stub</artifactId>
        <version>1.60.0</version>
    </dependency>

    <!-- JSON processing -->
    <dependency>
        <groupId>com.fasterxml.jackson.core</groupId>
        <artifactId>jackson-databind</artifactId>
        <version>2.15.0</version>
    </dependency>

    <!-- Logging -->
    <dependency>
        <groupId>org.slf4j</groupId>
        <artifactId>slf4j-api</artifactId>
        <version>2.0.7</version>
    </dependency>
    <dependency>
        <groupId>ch.qos.logback</groupId>
        <artifactId>logback-classic</artifactId>
        <version>1.4.8</version>
    </dependency>
</dependencies>
```

## Quick Start

```java
import com.deepapp.sdk.Worker;
import com.deepapp.sdk.Capability;
import com.deepapp.sdk.CapabilityHandler;

public class MyWorker extends Worker {

    public MyWorker() {
        super("my-java-worker", "localhost:50051");
    }

    @Override
    protected List<Capability> getCapabilities() {
        return Arrays.asList(
            new Capability(
                "hello",
                "Returns a hello message",
                "{}",
                "{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}",
                "GET",
                false,
                null
            ),
            new Capability(
                "process_data",
                "Process uploaded data",
                "{\"type\":\"object\",\"properties\":{\"file\":{\"type\":\"string\",\"format\":\"binary\"}}}",
                "{\"type\":\"object\",\"properties\":{\"result\":{\"type\":\"string\"}}}",
                "POST",
                true,
                "file"
            )
        );
    }

    @CapabilityHandler("hello")
    public String handleHello(String input) {
        return "{\"message\":\"Hello from Java Worker! ☕\",\"timestamp\":\"" +
               java.time.Instant.now() + "\",\"workerId\":\"" + getWorkerId() + "\"}";
    }

    @CapabilityHandler("process_data")
    public String handleProcessData(String input) {
        try {
            // Parse JSON input
            ObjectMapper mapper = new ObjectMapper();
            JsonNode data = mapper.readTree(input);
            String filename = data.get("filename").asText("unknown");

            return mapper.writeValueAsString(Map.of(
                "filename", filename,
                "processed", true,
                "result", "Data processed successfully",
                "timestamp", java.time.Instant.now().toString()
            ));
        } catch (Exception e) {
            return "{\"error\":\"" + e.getMessage() + "\",\"status\":\"failed\"}";
        }
    }

    public static void main(String[] args) {
        MyWorker worker = new MyWorker();
        worker.start();
    }
}
```

## API Reference

### Worker Class

#### Constructor

```java
public Worker(String workerId, String hubAddress)
```

#### Methods

- `void start()` - Connect to hub and start processing
- `void stop()` - Disconnect from hub
- `String getWorkerId()` - Get worker ID
- `List<Capability> getCapabilities()` - Override to define capabilities

### Capability Class

```java
public Capability(
    String name,           // Unique capability name
    String description,    // Human readable description
    String inputSchema,    // JSON Schema for input
    String outputSchema,   // JSON Schema for output
    String httpMethod,     // HTTP method for web API
    boolean acceptsFile,   // Whether it accepts file uploads
    String fileFieldName   // Field name for file uploads
)
```

### CapabilityHandler Annotation

```java
@CapabilityHandler("capability_name")
public String handleCapability(String input) {
    // Your logic here
    return "{\"result\":\"success\"}";
}
```

## Advanced Usage

### File Upload Handling

```java
@CapabilityHandler("analyze_image")
public String handleAnalyzeImage(String input) {
    try {
        ObjectMapper mapper = new ObjectMapper();
        JsonNode data = mapper.readTree(input);

        // File data is base64 encoded
        String fileData = data.get("file").asText();
        byte[] fileBytes = Base64.getDecoder().decode(fileData);

        // Process the file...
        String analysis = analyzeImage(fileBytes);

        return mapper.writeValueAsString(Map.of(
            "analysis", analysis,
            "timestamp", java.time.Instant.now().toString()
        ));
    } catch (Exception e) {
        return "{\"error\":\"" + e.getMessage() + "\",\"status\":\"failed\"}";
    }
}
```

### Error Handling

```java
@CapabilityHandler("process_data")
public String handleProcessData(String input) {
    try {
        // Your processing logic
        return "{\"result\":\"success\"}";
    } catch (Exception e) {
        // Return error response
        return "{\"error\":\"" + e.getMessage() + "\",\"status\":\"failed\"}";
    }
}
```

### Environment Variables

```bash
WORKER_ID=my-custom-worker
HUB_ADDRESS=localhost:50051
```

## Complete Example

See `examples/` directory for complete working examples with Maven project structure.