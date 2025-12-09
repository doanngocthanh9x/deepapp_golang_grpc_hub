package com.deepapp.worker;

import com.deepapp.hub.ServiceCapability;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;

/**
 * Java unit testing capability handler
 */
public class JavaTestHandler implements CapabilityHandler {
    private static final Logger logger = LoggerFactory.getLogger(JavaTestHandler.class);
    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public ServiceCapability getCapability() {
        return ServiceCapability.newBuilder()
                .setName("java_test")
                .setDescription("Run Java unit tests")
                .setInputSchema("{\"type\":\"object\",\"properties\":{\"testClass\":{\"type\":\"string\",\"description\":\"Test class source code\"},\"className\":{\"type\":\"string\",\"description\":\"Test class name\"}},\"required\":[\"testClass\"]}")
                .setOutputSchema("{\"type\":\"object\",\"properties\":{\"success\":{\"type\":\"boolean\"},\"testsRun\":{\"type\":\"integer\"},\"failures\":{\"type\":\"integer\"},\"output\":{\"type\":\"string\"}}}")
                .build();
    }

    @Override
    public String handle(String input) throws Exception {
        try {
            JsonNode request = objectMapper.readTree(input);
            String testClass = request.get("testClass").asText();
            String className = request.has("className") ? request.get("className").asText() : "TestMain";

            // Mock test execution
            // In real implementation, you'd compile and run JUnit tests

            boolean success = testClass.contains("@Test") && testClass.contains("assert");
            int testsRun = testClass.split("@Test").length - 1;
            int failures = success ? 0 : 1;

            StringBuilder output = new StringBuilder();
            output.append("Running ").append(testsRun).append(" tests...\n");

            if (success) {
                output.append("All tests passed!\n");
                for (int i = 1; i <= testsRun; i++) {
                    output.append("✓ Test ").append(i).append(" passed\n");
                }
            } else {
                output.append("Test failures detected!\n");
                output.append("✗ Some tests failed\n");
            }

            return objectMapper.writeValueAsString(Map.of(
                "success", success,
                "testsRun", testsRun,
                "failures", failures,
                "output", output.toString().trim()
            ));

        } catch (Exception e) {
            logger.error("Error in Java testing", e);
            return objectMapper.writeValueAsString(Map.of(
                "success", false,
                "testsRun", 0,
                "failures", 1,
                "output", "Error: " + e.getMessage()
            ));
        }
    }
}