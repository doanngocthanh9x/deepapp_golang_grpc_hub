package com.deepapp.worker;

import com.deepapp.hub.ServiceCapability;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Arrays;
import java.util.Map;

/**
 * Maven build capability handler
 */
public class MavenBuildHandler implements CapabilityHandler {
    private static final Logger logger = LoggerFactory.getLogger(MavenBuildHandler.class);
    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public ServiceCapability getCapability() {
        return ServiceCapability.newBuilder()
                .setName("maven_build")
                .setDescription("Execute Maven build commands")
                .setInputSchema("{\"type\":\"object\",\"properties\":{\"goals\":{\"type\":\"array\",\"items\":{\"type\":\"string\"},\"description\":\"Maven goals to execute\",\"default\":[\"clean\",\"compile\"]},\"pom\":{\"type\":\"string\",\"description\":\"Custom pom.xml content\"}},\"required\":[]}")
                .setOutputSchema("{\"type\":\"object\",\"properties\":{\"success\":{\"type\":\"boolean\"},\"output\":{\"type\":\"string\"},\"exitCode\":{\"type\":\"integer\"}}}")
                .build();
    }

    @Override
    public String handle(String input) throws Exception {
        try {
            JsonNode request = objectMapper.readTree(input);

            // Default goals
            String[] goals = {"clean", "compile"};
            if (request.has("goals")) {
                JsonNode goalsNode = request.get("goals");
                goals = new String[goalsNode.size()];
                for (int i = 0; i < goalsNode.size(); i++) {
                    goals[i] = goalsNode.get(i).asText();
                }
            }

            // Create temporary directory
            Path tempDir = Files.createTempDirectory("maven_build_");

            try {
                // Create pom.xml if provided
                if (request.has("pom")) {
                    String pomContent = request.get("pom").asText();
                    Path pomFile = tempDir.resolve("pom.xml");
                    Files.writeString(pomFile, pomContent);
                } else {
                    // Create basic pom.xml
                    createBasicPom(tempDir);
                }

                // Execute Maven
                ProcessBuilder pb = new ProcessBuilder("mvn");
                pb.command().addAll(Arrays.asList(goals));
                pb.directory(tempDir.toFile());
                pb.redirectErrorStream(true);

                Process process = pb.start();

                // Read output
                StringBuilder output = new StringBuilder();
                try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                    String line;
                    while ((line = reader.readLine()) != null) {
                        output.append(line).append("\n");
                    }
                }

                int exitCode = process.waitFor();

                boolean success = exitCode == 0;

                return objectMapper.writeValueAsString(Map.of(
                    "success", success,
                    "output", output.toString().trim(),
                    "exitCode", exitCode
                ));

            } finally {
                // Cleanup
                deleteDirectory(tempDir);
            }

        } catch (Exception e) {
            logger.error("Error in Maven build", e);
            return objectMapper.writeValueAsString(Map.of(
                "success", false,
                "output", e.getMessage(),
                "exitCode", -1
            ));
        }
    }

    private void createBasicPom(Path dir) throws IOException {
        String pomContent = """
            <?xml version="1.0" encoding="UTF-8"?>
            <project xmlns="http://maven.apache.org/POM/4.0.0"
                     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                     xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
                     http://maven.apache.org/xsd/maven-4.0.0.xsd">
                <modelVersion>4.0.0</modelVersion>
                <groupId>com.example</groupId>
                <artifactId>test-project</artifactId>
                <version>1.0.0</version>
                <properties>
                    <maven.compiler.source>17</maven.compiler.source>
                    <maven.compiler.target>17</maven.compiler.target>
                </properties>
                <dependencies>
                    <dependency>
                        <groupId>junit</groupId>
                        <artifactId>junit</artifactId>
                        <version>4.13.2</version>
                        <scope>test</scope>
                    </dependency>
                </dependencies>
            </project>
            """;

        Files.writeString(dir.resolve("pom.xml"), pomContent.trim());
    }

    private void deleteDirectory(Path path) {
        try {
            Files.walk(path)
                .sorted((a, b) -> b.compareTo(a))
                .forEach(p -> {
                    try {
                        Files.delete(p);
                    } catch (IOException e) {
                        logger.warn("Failed to delete: {}", p, e);
                    }
                });
        } catch (IOException e) {
            logger.warn("Failed to delete directory: {}", path, e);
        }
    }
}