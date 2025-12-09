package com.deepapp.examples;

import com.deepapp.sdk.Worker;
import com.deepapp.sdk.Capability;
import com.deepapp.sdk.CapabilityHandler;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.util.Arrays;
import java.util.List;
import java.util.Map;

/**
 * Example worker demonstrating basic capabilities
 */
public class SimpleWorker extends Worker {

    private final ObjectMapper objectMapper;

    public SimpleWorker() {
        super(
            System.getenv().getOrDefault("WORKER_ID", "java-example-worker"),
            System.getenv().getOrDefault("HUB_ADDRESS", "localhost:50051")
        );
        this.objectMapper = new ObjectMapper();
    }

    @Override
    protected List<Capability> getCapabilities() {
        return Arrays.asList(
            new Capability(
                "hello",
                "Returns a hello message",
                "{}",
                "{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"},\"timestamp\":{\"type\":\"string\"},\"workerId\":{\"type\":\"string\"}}}",
                "GET",
                false,
                null
            ),
            new Capability(
                "echo",
                "Echoes back the input message",
                "{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}",
                "{\"type\":\"object\",\"properties\":{\"echo\":{\"type\":\"string\"},\"timestamp\":{\"type\":\"string\"}}}",
                "POST",
                false,
                null
            ),
            new Capability(
                "process_file",
                "Process an uploaded file",
                "{\"type\":\"object\",\"properties\":{\"file\":{\"type\":\"string\",\"format\":\"binary\"},\"filename\":{\"type\":\"string\"}}}",
                "{\"type\":\"object\",\"properties\":{\"filename\":{\"type\":\"string\"},\"size\":{\"type\":\"number\"},\"processed\":{\"type\":\"boolean\"},\"timestamp\":{\"type\":\"string\"}}}",
                "POST",
                true,
                "file"
            )
        );
    }

    @CapabilityHandler("hello")
    public String handleHello(String input) {
        try {
            return objectMapper.writeValueAsString(Map.of(
                "message", "Hello World from Java Worker! ☕",
                "timestamp", java.time.Instant.now().toString(),
                "workerId", getWorkerId(),
                "status", "success"
            ));
        } catch (Exception e) {
            return "{\"error\":\"" + e.getMessage() + "\",\"status\":\"failed\"}";
        }
    }

    @CapabilityHandler("echo")
    public String handleEcho(String input) {
        try {
            // Parse input JSON
            Map<String, Object> data = objectMapper.readValue(input, Map.class);
            String message = (String) data.getOrDefault("message", "No message provided");

            return objectMapper.writeValueAsString(Map.of(
                "echo", message,
                "timestamp", java.time.Instant.now().toString(),
                "status", "success"
            ));
        } catch (Exception e) {
            return "{\"error\":\"" + e.getMessage() + "\",\"status\":\"failed\"}";
        }
    }

    @CapabilityHandler("process_file")
    public String handleProcessFile(String input) {
        try {
            // Parse input JSON
            Map<String, Object> data = objectMapper.readValue(input, Map.class);
            String filename = (String) data.getOrDefault("filename", "unknown");
            String fileData = (String) data.get("file");

            if (fileData == null) {
                return objectMapper.writeValueAsString(Map.of(
                    "error", "No file data provided",
                    "status", "failed"
                ));
            }

            // Decode base64 file data
            byte[] fileBytes = java.util.Base64.getDecoder().decode(fileData);
            int fileSize = fileBytes.length;

            // Simulate file processing
            System.out.println("📁 Processing file: " + filename + " (" + fileSize + " bytes)");

            // Here you would do actual file processing
            // For example: image analysis, text extraction, etc.

            return objectMapper.writeValueAsString(Map.of(
                "filename", filename,
                "size", fileSize,
                "processed", true,
                "result", "File processed successfully",
                "timestamp", java.time.Instant.now().toString(),
                "status", "success"
            ));
        } catch (Exception e) {
            System.err.println("Error processing file: " + e.getMessage());
            try {
                return objectMapper.writeValueAsString(Map.of(
                    "error", e.getMessage(),
                    "status", "failed"
                ));
            } catch (Exception ex) {
                return "{\"error\":\"JSON serialization failed\",\"status\":\"failed\"}";
            }
        }
    }

    public static void main(String[] args) {
        SimpleWorker worker = new SimpleWorker();

        // Add shutdown hook for graceful shutdown
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            System.out.println("\n🛑 Shutting down worker...");
            worker.stop();
        }));

        try {
            worker.start();
        } catch (Exception e) {
            System.err.println("Failed to start worker: " + e.getMessage());
            System.exit(1);
        }
    }
}