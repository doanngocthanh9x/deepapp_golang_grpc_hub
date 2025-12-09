package com.deepapp.worker;

import com.deepapp.hub.Hub.Message;
import com.deepapp.hub.Hub.MessageType;
import com.deepapp.hub.Hub.ServiceCapability;
import com.deepapp.hub.Hub.WorkerRegistration;
import com.deepapp.hub.HubServiceGrpc;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

/**
 * Simple Java Worker for gRPC Hub System
 * Provides basic capabilities: hello_world and read_file_info
 */
public class SimpleWorker {
    private static final Logger logger = LoggerFactory.getLogger(SimpleWorker.class);

    private final String workerId;
    private final String hubAddress;
    private final ManagedChannel channel;
    private final HubServiceGrpc.HubServiceStub asyncStub;
    private final Map<String, CapabilityHandler> handlers;
    private final ObjectMapper objectMapper;
    private StreamObserver<Message> requestObserver;
    private final Map<String, PendingCall> pendingCalls;  // Track worker-to-worker calls

    // Helper class for pending calls
    private static class PendingCall {
        final Object lock = new Object();
        Message response = null;
        boolean completed = false;
    }

    public SimpleWorker(String workerId, String hubAddress) {
        this.workerId = workerId;
        this.hubAddress = hubAddress;

        try {
            // Parse host and port from hubAddress (format: host:port)
            String[] parts = hubAddress.split(":");
            String host = parts[0];
            int port = Integer.parseInt(parts[1]);

            logger.info("Connecting to Hub at {}:{}", host, port);

            // Create gRPC channel with OkHttp transport
            this.channel = ManagedChannelBuilder.forAddress(host, port)
                    .usePlaintext()
                    .build();

            logger.info("Channel created successfully");
        } catch (Exception e) {
            logger.error("Failed to create gRPC channel", e);
            throw new RuntimeException("Failed to create gRPC channel", e);
        }

        this.asyncStub = HubServiceGrpc.newStub(channel);
        this.handlers = new ConcurrentHashMap<>();
        this.objectMapper = new ObjectMapper();
        this.pendingCalls = new ConcurrentHashMap<>();

        // Register capabilities
        registerCapabilities();
    }

    private void registerCapabilities() {
        // Hello World capability
        handlers.put("hello_world", new HelloWorldHandler());

        // File info capability
        handlers.put("read_file_info", new FileInfoHandler());

        logger.info("Registered {} capabilities: {}", handlers.size(), handlers.keySet());
    }

    public void start() {
        logger.info("Starting Simple Worker {} connecting to {}", workerId, hubAddress);

        this.requestObserver = asyncStub.connect(new StreamObserver<Message>() {
            @Override
            public void onNext(Message message) {
                handleMessage(message);
            }

            @Override
            public void onError(Throwable t) {
                logger.error("Stream error", t);
            }

            @Override
            public void onCompleted() {
                logger.info("Stream completed");
            }
        });

        // Send registration message
        sendRegistration();

        // Keep the connection alive
        try {
            Thread.currentThread().join();
        } catch (InterruptedException e) {
            logger.info("Worker interrupted, shutting down");
            shutdown();
        }
    }

    private void sendRegistration() {
        try {
            List<ServiceCapability> capabilities = new ArrayList<>();
            for (CapabilityHandler handler : handlers.values()) {
                capabilities.add(handler.getCapability());
            }

            // Build registration data as JSON
            Map<String, Object> registrationData = new HashMap<>();
            registrationData.put("worker_id", workerId);
            registrationData.put("worker_type", "java-simple");
            
            // Convert capabilities to JSON-friendly format
            List<Map<String, Object>> capsList = new ArrayList<>();
            for (ServiceCapability cap : capabilities) {
                Map<String, Object> capMap = new HashMap<>();
                capMap.put("name", cap.getName());
                capMap.put("description", cap.getDescription());
                capMap.put("input_schema", cap.getInputSchema());
                capMap.put("output_schema", cap.getOutputSchema());
                
                // Add HTTP method and file handling info
                if ("read_file_info".equals(cap.getName())) {
                    capMap.put("http_method", "POST");
                    capMap.put("accepts_file", false);
                } else {
                    capMap.put("http_method", "POST");
                    capMap.put("accepts_file", false);
                }
                
                capsList.add(capMap);
            }
            registrationData.put("capabilities", capsList);
            
            Map<String, String> metadata = new HashMap<>();
            metadata.put("version", "1.0.0");
            metadata.put("description", "Simple Java worker with basic capabilities");
            registrationData.put("metadata", metadata);

            String jsonContent = objectMapper.writeValueAsString(registrationData);

            Message regMessage = Message.newBuilder()
                    .setId(UUID.randomUUID().toString())
                    .setFrom(workerId)
                    .setTo("hub")
                    .setChannel("system")
                    .setContent(jsonContent)
                    .setTimestamp(String.valueOf(System.currentTimeMillis()))
                    .setType(MessageType.REGISTER)
                    .setAction("register")
                    .putMetadata("registration", "true")
                    .build();

            requestObserver.onNext(regMessage);
            logger.info("Sent registration for worker {} with {} capabilities", workerId, capabilities.size());
        } catch (Exception e) {
            logger.error("Failed to send registration", e);
        }
    }

    /**
     * Call another worker's capability through Hub
     * 
     * @param targetWorker Target worker ID (e.g., "python-worker")
     * @param capability Capability name to call
     * @param params Parameters as JSON string
     * @param timeoutSeconds Timeout in seconds
     * @return Response message from the target worker
     * @throws Exception if call fails or times out
     */
    public Message callWorker(String targetWorker, String capability, String params, int timeoutSeconds) throws Exception {
        String requestId = UUID.randomUUID().toString();
        
        logger.info("🔗 Calling worker '{}' capability '{}'", targetWorker, capability);
        
        // Create worker call message
        Message callMessage = Message.newBuilder()
                .setId(requestId)
                .setFrom(workerId)
                .setTo(targetWorker)
                .setChannel(capability)
                .setContent(params)
                .setTimestamp(String.valueOf(System.currentTimeMillis()))
                .setType(MessageType.WORKER_CALL)
                .putMetadata("capability", capability)
                .build();
        
        // Register pending call
        PendingCall pending = new PendingCall();
        pendingCalls.put(requestId, pending);
        
        try {
            // Send the call
            requestObserver.onNext(callMessage);
            
            // Wait for response
            synchronized (pending.lock) {
                long startTime = System.currentTimeMillis();
                long timeoutMillis = timeoutSeconds * 1000L;
                
                while (!pending.completed) {
                    long elapsed = System.currentTimeMillis() - startTime;
                    long remaining = timeoutMillis - elapsed;
                    
                    if (remaining <= 0) {
                        throw new Exception("Timeout waiting for response from " + targetWorker);
                    }
                    
                    pending.lock.wait(remaining);
                }
            }
            
            if (pending.response == null) {
                throw new Exception("No response received from " + targetWorker);
            }
            
            logger.info("✓ Received response from {}", targetWorker);
            return pending.response;
            
        } finally {
            pendingCalls.remove(requestId);
        }
    }

    private void handleMessage(Message message) {
        logger.info("📨 Received message: action='{}', from={}, to={}, type={}, metadata={}", 
                message.getAction(), message.getFrom(), message.getTo(), 
                message.getType(), message.getMetadataMap());

        // Handle WORKER_CALL from another worker
        if (message.getType() == MessageType.WORKER_CALL) {
            logger.info("🔗 Worker-to-Worker call from {}, capability: {}", 
                    message.getFrom(), message.getChannel());
            handleWorkerCall(message);
            return;
        }

        // Handle RESPONSE (could be from worker-to-worker call)
        if (message.getType() == MessageType.RESPONSE) {
            logger.info("📬 Response from {}", message.getFrom());
            handleResponse(message);
            return;
        }

        // Handle regular REQUEST
        if ("request".equals(message.getAction()) && message.getMetadataMap().containsKey("capability")) {
            String capability = message.getMetadataMap().get("capability");
            logger.info("🔍 Processing capability: {}", capability);
            if (handlers.containsKey(capability)) {
                handleCapabilityRequest(message, capability);
            } else {
                logger.warn("Unknown capability requested: {}", capability);
            }
        } else {
            logger.warn("Invalid request format - action='{}', has capability metadata={}", 
                    message.getAction(), message.getMetadataMap().containsKey("capability"));
        }
    }

    private void handleWorkerCall(Message message) {
        // Another worker is calling this worker
        String capability = message.getChannel();
        
        if (!handlers.containsKey(capability)) {
            logger.warn("Worker call for unknown capability: {}", capability);
            sendErrorResponse(message, "Capability not found: " + capability);
            return;
        }

        try {
            CapabilityHandler handler = handlers.get(capability);
            String input = message.getContent();
            String result = handler.handle(input);

            // Send response back to calling worker
            Message response = Message.newBuilder()
                    .setId(UUID.randomUUID().toString())
                    .setFrom(workerId)
                    .setTo(message.getFrom())  // Send back to caller
                    .setChannel(capability)
                    .setContent(result)
                    .setTimestamp(String.valueOf(System.currentTimeMillis()))
                    .setType(MessageType.RESPONSE)
                    .putMetadata("request_id", message.getId())
                    .putMetadata("status", "success")
                    .build();

            requestObserver.onNext(response);
            logger.info("✓ Sent worker call response to {}", message.getFrom());

        } catch (Exception e) {
            logger.error("Error handling worker call: {}", e.getMessage(), e);
            sendErrorResponse(message, "Error: " + e.getMessage());
        }
    }

    private void handleResponse(Message message) {
        // Check if this is a response to our pending call
        String requestId = message.getMetadataMap().get("request_id");
        
        if (requestId != null && pendingCalls.containsKey(requestId)) {
            PendingCall pending = pendingCalls.get(requestId);
            synchronized (pending.lock) {
                pending.response = message;
                pending.completed = true;
                pending.lock.notifyAll();
            }
            logger.info("✓ Matched response to pending call {}", requestId);
        } else {
            logger.debug("Received response without matching pending call");
        }
    }

    private void sendErrorResponse(Message originalMessage, String errorMessage) {
        try {
            Map<String, String> errorData = new HashMap<>();
            errorData.put("error", errorMessage);
            errorData.put("status", "failed");
            
            String jsonError = objectMapper.writeValueAsString(errorData);
            
            Message response = Message.newBuilder()
                    .setId(UUID.randomUUID().toString())
                    .setFrom(workerId)
                    .setTo(originalMessage.getFrom())
                    .setChannel(originalMessage.getChannel())
                    .setContent(jsonError)
                    .setTimestamp(String.valueOf(System.currentTimeMillis()))
                    .setType(MessageType.RESPONSE)
                    .putMetadata("request_id", originalMessage.getId())
                    .putMetadata("status", "error")
                    .build();

            requestObserver.onNext(response);
        } catch (Exception e) {
            logger.error("Failed to send error response", e);
        }
    }

    private void handleCapabilityRequest(Message request, String capability) {
        try {
            CapabilityHandler handler = handlers.get(capability);
            String input = request.getContent();
            String result = handler.handle(input);

            // Send response back
            Message response = Message.newBuilder()
                    .setId(UUID.randomUUID().toString())
                    .setFrom(workerId)
                    .setTo(request.getFrom())
                    .setChannel(request.getChannel())
                    .setContent(result)
                    .setTimestamp(String.valueOf(System.currentTimeMillis()))
                    .setType(MessageType.RESPONSE)
                    .setAction("response")
                    .putMetadata("request_id", request.getId())
                    .putMetadata("status", "success")
                    .putMetadata("capability", capability)
                    .build();

            // Send response back
            requestObserver.onNext(response);
            logger.info("Capability {} executed successfully: {}", capability, result);

        } catch (Exception e) {
            logger.error("Error handling capability {}: {}", capability, e.getMessage(), e);
        }
    }

    public void shutdown() {
        try {
            channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
        } catch (InterruptedException e) {
            logger.warn("Channel shutdown interrupted", e);
        }
        logger.info("Worker {} shut down", workerId);
    }

    public static void main(String[] args) {
        String workerId = System.getenv("WORKER_ID");
        if (workerId == null || workerId.isEmpty()) {
            workerId = "java-simple-worker-" + System.currentTimeMillis();
        }

        String hubAddress = System.getenv("HUB_ADDRESS");
        if (hubAddress == null || hubAddress.isEmpty()) {
            hubAddress = "localhost:50051";
        }

        SimpleWorker worker = new SimpleWorker(workerId, hubAddress);

        // Add shutdown hook
        Runtime.getRuntime().addShutdownHook(new Thread(worker::shutdown));

        worker.start();
    }
}