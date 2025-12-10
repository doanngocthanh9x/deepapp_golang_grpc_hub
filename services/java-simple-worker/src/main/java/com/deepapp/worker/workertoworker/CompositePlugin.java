package com.deepapp.worker.workertoworker;

import com.deepapp.worker.plugins.BasePlugin;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import java.util.function.BiFunction;

/**
 * Composite Plugin - Demonstrates Java worker calling other workers
 */
public class CompositePlugin implements BasePlugin {
    private static final ObjectMapper mapper = new ObjectMapper();

    @Override
    public String getName() {
        return "java_composite";
    }

    @Override
    public String getDescription() {
        return "Calls multiple workers (Python, Node.js, Go) and combines their results";
    }

    @Override
    public String execute(String requestData, Object workerSDK) throws Exception {
        System.out.println("🔄 Starting composite task from Java...");

        // Cast workerSDK to Map
        @SuppressWarnings("unchecked")
        Map<String, Object> context = (Map<String, Object>) workerSDK;

        ObjectNode results = mapper.createObjectNode();
        results.put("worker_id", context.get("worker_id").toString());
        results.put("timestamp", System.currentTimeMillis());

        ArrayNode calls = mapper.createArrayNode();

        // Get callWorker function from context
        @SuppressWarnings("unchecked")
        BiFunction<String, String, CompletableFuture<String>> callWorker =
                (BiFunction<String, String, CompletableFuture<String>>) context.get("callWorker");

        // Call Python worker
        try {
            System.out.println("  → Calling Python worker (hello)...");
            ObjectNode params = mapper.createObjectNode();
            params.put("name", "from Java");

            String pythonResponse = callWorker
                    .apply("python-worker:hello", mapper.writeValueAsString(params))
                    .get(10, TimeUnit.SECONDS);

            ObjectNode call = mapper.createObjectNode();
            call.put("worker", "python-worker");
            call.put("capability", "hello");
            call.put("status", "success");
            call.set("response", mapper.readTree(pythonResponse));
            calls.add(call);

        } catch (Exception e) {
            System.out.println("  ✗ Python worker call failed: " + e.getMessage());
            ObjectNode call = mapper.createObjectNode();
            call.put("worker", "python-worker");
            call.put("capability", "hello");
            call.put("status", "error");
            call.put("error", e.getMessage());
            calls.add(call);
        }

        // Call Node.js worker
        try {
            System.out.println("  → Calling Node.js worker (hello_node)...");
            ObjectNode params = mapper.createObjectNode();
            params.put("name", "from Java");

            String nodeResponse = callWorker
                    .apply("node-worker:hello_node", mapper.writeValueAsString(params))
                    .get(10, TimeUnit.SECONDS);

            ObjectNode call = mapper.createObjectNode();
            call.put("worker", "node-worker");
            call.put("capability", "hello_node");
            call.put("status", "success");
            call.set("response", mapper.readTree(nodeResponse));
            calls.add(call);

        } catch (Exception e) {
            System.out.println("  ✗ Node.js worker call failed: " + e.getMessage());
            ObjectNode call = mapper.createObjectNode();
            call.put("worker", "node-worker");
            call.put("capability", "hello_node");
            call.put("status", "error");
            call.put("error", e.getMessage());
            calls.add(call);
        }

        // Call Go worker
        try {
            System.out.println("  → Calling Go worker (hello_go)...");
            ObjectNode params = mapper.createObjectNode();
            params.put("name", "from Java");

            String goResponse = callWorker
                    .apply("go-worker:hello_go", mapper.writeValueAsString(params))
                    .get(10, TimeUnit.SECONDS);

            ObjectNode call = mapper.createObjectNode();
            call.put("worker", "go-worker");
            call.put("capability", "hello_go");
            call.put("status", "success");
            call.set("response", mapper.readTree(goResponse));
            calls.add(call);

        } catch (Exception e) {
            System.out.println("  ✗ Go worker call failed: " + e.getMessage());
            ObjectNode call = mapper.createObjectNode();
            call.put("worker", "go-worker");
            call.put("capability", "hello_go");
            call.put("status", "error");
            call.put("error", e.getMessage());
            calls.add(call);
        }

        results.set("calls", calls);

        // Summary
        int successCount = 0;
        for (int i = 0; i < calls.size(); i++) {
            if ("success".equals(calls.get(i).get("status").asText())) {
                successCount++;
            }
        }

        ObjectNode summary = mapper.createObjectNode();
        summary.put("total_calls", calls.size());
        summary.put("successful", successCount);
        summary.put("failed", calls.size() - successCount);
        results.set("summary", summary);

        System.out.println("✅ Composite task completed");
        return mapper.writeValueAsString(results);
    }
}
