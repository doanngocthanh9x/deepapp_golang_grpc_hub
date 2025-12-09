package com.deepapp.sdk;

import com.deepapp.hub.Hub.Message;
import com.deepapp.hub.Hub.MessageType;
import com.deepapp.hub.HubServiceGrpc;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.lang.reflect.Method;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

/**
 * Base class for creating workers that connect to the DeepApp gRPC Hub
 */
public abstract class Worker {
    private static final Logger logger = LoggerFactory.getLogger(Worker.class);

    protected final String workerId;
    protected final String hubAddress;
    private final ManagedChannel channel;
    private final HubServiceGrpc.HubServiceStub asyncStub;
    private final Map<String, Method> handlers;
    private final ObjectMapper objectMapper;
    private StreamObserver<Message> requestObserver;
    private volatile boolean running = false;

    public Worker(String workerId, String hubAddress) {
        this.workerId = workerId;
        this.hubAddress = hubAddress;
        this.handlers = new ConcurrentHashMap<>();
        this.objectMapper = new ObjectMapper();

        try {
            // Parse host and port
            String[] parts = hubAddress.split(":");
            String host = parts[0];
            int port = Integer.parseInt(parts[1]);

            logger.info("Connecting to Hub at {}:{}", host, port);

            // Create gRPC channel with OkHttp transport (critical for Unix socket support)
            this.channel = ManagedChannelBuilder.forAddress(host, port)
                    .usePlaintext()
                    .build();

            logger.info("Channel created successfully");
        } catch (Exception e) {
            logger.error("Failed to create gRPC channel", e);
            throw new RuntimeException("Failed to create gRPC channel", e);
        }

        this.asyncStub = HubServiceGrpc.newStub(channel);

        // Register capability handlers
        registerHandlers();
    }

    /**
     * Override this method to define your worker's capabilities
     */
    protected abstract List<Capability> getCapabilities();

    /**
     * Register capability handler methods
     */
    private void registerHandlers() {
        Method[] methods = this.getClass().getMethods();
        for (Method method : methods) {
            CapabilityHandler annotation = method.getAnnotation(CapabilityHandler.class);
            if (annotation != null) {
                String capabilityName = annotation.value();
                handlers.put(capabilityName, method);
                logger.info("Registered handler for capability: {}", capabilityName);
            }
        }
    }

    /**
     * Start the worker
     */
    public void start() {
        logger.info("Starting Worker {} connecting to {}", workerId, hubAddress);

        this.requestObserver = asyncStub.connect(new StreamObserver<Message>() {
            @Override
            public void onNext(Message message) {
                handleMessage(message);
            }

            @Override
            public void onError(Throwable t) {
                logger.error("Stream error", t);
                running = false;
            }

            @Override
            public void onCompleted() {
                logger.info("Stream completed");
                running = false;
            }
        });

        running = true;

        // Send registration message
        sendRegistration();

        // Keep the connection alive
        try {
            while (running) {
                Thread.sleep(1000);
            }
        } catch (InterruptedException e) {
            logger.info("Worker interrupted, shutting down");
            stop();
        }
    }

    /**
     * Stop the worker
     */
    public void stop() {
        logger.info("Stopping worker {}", workerId);
        running = false;

        if (requestObserver != null) {
            requestObserver.onCompleted();
        }

        try {
            channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
        } catch (InterruptedException e) {
            logger.warn("Channel shutdown interrupted", e);
        }

        logger.info("Worker {} stopped", workerId);
    }

    /**
     * Send worker registration to the hub
     */
    private void sendRegistration() {
        try {
            List<Capability> capabilities = getCapabilities();

            // Convert capabilities to JSON-friendly format
            List<Map<String, Object>> capsList = new ArrayList<>();
            for (Capability cap : capabilities) {
                Map<String, Object> capMap = new HashMap<>();
                capMap.put("name", cap.getName());
                capMap.put("description", cap.getDescription());
                capMap.put("input_schema", cap.getInputSchema());
                capMap.put("output_schema", cap.getOutputSchema());
                capMap.put("http_method", cap.getHttpMethod());
                capMap.put("accepts_file", cap.isAcceptsFile());
                if (cap.getFileFieldName() != null) {
                    capMap.put("file_field_name", cap.getFileFieldName());
                }
                capsList.add(capMap);
            }

            // Build registration data
            Map<String, Object> registrationData = new HashMap<>();
            registrationData.put("worker_id", workerId);
            registrationData.put("worker_type", "java-sdk");
            registrationData.put("capabilities", capsList);

            Map<String, String> metadata = new HashMap<>();
            metadata.put("version", "1.0.0");
            metadata.put("description", "Java SDK Worker");
            metadata.put("sdk", "java-sdk");
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
     * Handle incoming messages
     */
    private void handleMessage(Message message) {
        logger.info("📨 Received message: action='{}', from={}, channel={}",
                message.getAction(), message.getFrom(), message.getChannel());

        // Check if this is a capability request
        String capability = message.getChannel();
        if (handlers.containsKey(capability)) {
            handleCapabilityRequest(message, capability);
        } else {
            logger.warn("Unknown capability requested: {}", capability);
        }
    }

    /**
     * Handle capability execution request
     */
    private void handleCapabilityRequest(Message request, String capability) {
        try {
            Method handler = handlers.get(capability);
            String input = request.getContent();

            logger.info("🔍 Executing capability: {}", capability);

            // Invoke the handler method
            String result = (String) handler.invoke(this, input);

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

            requestObserver.onNext(response);
            logger.info("✅ Capability {} executed successfully", capability);

        } catch (Exception e) {
            logger.error("❌ Error handling capability {}: {}", capability, e.getMessage(), e);

            // Send error response
            try {
                String errorResponse = objectMapper.writeValueAsString(Map.of(
                    "error", e.getMessage(),
                    "status", "failed",
                    "capability", capability
                ));

                Message errorMsg = Message.newBuilder()
                        .setId(UUID.randomUUID().toString())
                        .setFrom(workerId)
                        .setTo(request.getFrom())
                        .setChannel(request.getChannel())
                        .setContent(errorResponse)
                        .setTimestamp(String.valueOf(System.currentTimeMillis()))
                        .setType(MessageType.RESPONSE)
                        .setAction("response")
                        .putMetadata("request_id", request.getId())
                        .putMetadata("status", "error")
                        .putMetadata("capability", capability)
                        .build();

                requestObserver.onNext(errorMsg);
            } catch (Exception ex) {
                logger.error("Failed to send error response", ex);
            }
        }
    }

    // Getters
    public String getWorkerId() { return workerId; }
    public String getHubAddress() { return hubAddress; }
    public boolean isRunning() { return running; }
}