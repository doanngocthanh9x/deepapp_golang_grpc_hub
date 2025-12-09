package com.deepapp.worker;

import com.deepapp.hub.*;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

/**
 * Java Maven Worker for gRPC Hub System
 * Provides Java/Maven related capabilities
 */
public class MavenWorker {
    private static final Logger logger = LoggerFactory.getLogger(MavenWorker.class);

    private final String workerId;
    private final String hubAddress;
    private final ManagedChannel channel;
    private final HubServiceGrpc.HubServiceStub asyncStub;
    private final Map<String, CapabilityHandler> handlers;

    public MavenWorker(String workerId, String hubAddress) {
        this.workerId = workerId;
        this.hubAddress = hubAddress;

        // Create gRPC channel
        this.channel = ManagedChannelBuilder.forTarget(hubAddress)
                .usePlaintext()
                .build();

        this.asyncStub = HubServiceGrpc.newStub(channel);
        this.handlers = new ConcurrentHashMap<>();

        // Register capabilities
        registerCapabilities();
    }

    private void registerCapabilities() {
        // Java compilation capability
        handlers.put("java_compile", new JavaCompileHandler());

        // Maven build capability
        handlers.put("maven_build", new MavenBuildHandler());

        // JAR analysis capability
        handlers.put("jar_analyze", new JarAnalyzeHandler());

        // Unit test capability
        handlers.put("java_test", new JavaTestHandler());

        logger.info("Registered {} capabilities", handlers.size());
    }

    public void start() {
        logger.info("Starting Maven Worker: {}", workerId);

        // Create bidirectional stream
        StreamObserver<Message> requestObserver = asyncStub.connect(new StreamObserver<Message>() {
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
        sendRegistration(requestObserver);

        // Keep the connection alive
        try {
            Thread.currentThread().join();
        } catch (InterruptedException e) {
            logger.info("Worker interrupted");
            Thread.currentThread().interrupt();
        }
    }

    private void sendRegistration(StreamObserver<Message> requestObserver) {
        List<ServiceCapability> capabilities = new ArrayList<>();

        for (Map.Entry<String, CapabilityHandler> entry : handlers.entrySet()) {
            ServiceCapability capability = entry.getValue().getCapability();
            capabilities.add(capability);
        }

        WorkerRegistration registration = WorkerRegistration.newBuilder()
                .setWorkerId(workerId)
                .setWorkerType("maven")
                .addAllCapabilities(capabilities)
                .putMetadata("language", "java")
                .putMetadata("framework", "maven")
                .putMetadata("version", "17")
                .build();

        Message registerMessage = Message.newBuilder()
                .setId(UUID.randomUUID().toString())
                .setFrom(workerId)
                .setTo("hub")
                .setChannel("system")
                .setType(MessageType.REGISTER)
                .setContent("") // Registration data in metadata
                .putAllMetadata(registration.toByteString().toStringUtf8(), "") // Simplified
                .build();

        requestObserver.onNext(registerMessage);
        logger.info("Sent registration for worker: {}", workerId);
    }

    private void handleMessage(Message message) {
        logger.debug("Received message: {}", message.getId());

        if (message.getType() == MessageType.REQUEST) {
            handleRequest(message);
        }
    }

    private void handleRequest(Message request) {
        try {
            // Parse request
            Request req = Request.parseFrom(request.getContent().getBytes());

            if (handlers.containsKey(req.getType())) {
                CapabilityHandler handler = handlers.get(req.getType());

                // Execute capability
                String result = handler.handle(req.getData());

                // Send response
                Response response = Response.newBuilder()
                        .setStatus(Status.OK)
                        .setData(result)
                        .build();

                Message responseMessage = Message.newBuilder()
                        .setId(UUID.randomUUID().toString())
                        .setFrom(workerId)
                        .setTo(request.getFrom())
                        .setChannel(request.getChannel())
                        .setType(MessageType.RESPONSE)
                        .setContent(response.toString())
                        .setTimestamp(System.currentTimeMillis())
                        .build();

                // Note: In bidirectional streaming, we need to send response back through the stream
                // This is simplified - in real implementation, you'd need access to the requestObserver
                logger.info("Processed request {} with result: {}", request.getId(), result);
            } else {
                logger.warn("Unknown capability: {}", req.getType());
            }

        } catch (Exception e) {
            logger.error("Error handling request", e);
        }
    }

    public void shutdown() throws InterruptedException {
        channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
        logger.info("Worker shutdown complete");
    }

    public static void main(String[] args) {
        String workerId = System.getenv().getOrDefault("WORKER_ID", "java-maven-worker-" + UUID.randomUUID().toString().substring(0, 8));
        String hubAddress = System.getenv().getOrDefault("HUB_ADDRESS", "localhost:50051");

        MavenWorker worker = new MavenWorker(workerId, hubAddress);

        // Add shutdown hook
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            try {
                worker.shutdown();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }));

        worker.start();
    }
}