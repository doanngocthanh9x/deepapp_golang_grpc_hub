package com.deepapp.worker;

import com.deepapp.hub.ServiceCapability;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;

/**
 * Java compilation capability handler
 */
public class JavaCompileHandler implements CapabilityHandler {
    private static final Logger logger = LoggerFactory.getLogger(JavaCompileHandler.class);
    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public ServiceCapability getCapability() {
        return ServiceCapability.newBuilder()
                .setName("java_compile")
                .setDescription("Compile Java source code")
                .setInputSchema("{\"type\":\"object\",\"properties\":{\"source\":{\"type\":\"string\",\"description\":\"Java source code\"},\"className\":{\"type\":\"string\",\"description\":\"Main class name\"}},\"required\":[\"source\"]}")
                .setOutputSchema("{\"type\":\"object\",\"properties\":{\"success\":{\"type\":\"boolean\"},\"output\":{\"type\":\"string\"},\"errors\":{\"type\":\"array\",\"items\":{\"type\":\"string\"}}}}")
                .build();
    }

    @Override
    public String handle(String input) throws Exception {
        try {
            JsonNode request = objectMapper.readTree(input);
            String sourceCode = request.get("source").asText();
            String className = request.has("className") ? request.get("className").asText() : "Main";

            // Create temporary directory
            Path tempDir = Files.createTempDirectory("java_compile_");

            try {
                // Write source file
                Path sourceFile = tempDir.resolve(className + ".java");
                Files.writeString(sourceFile, sourceCode);

                // Compile
                JavaCompiler compiler = ToolProvider.getSystemJavaCompiler();
                DiagnosticCollector<JavaFileObject> diagnostics = new DiagnosticCollector<>();

                StandardJavaFileManager fileManager = compiler.getStandardFileManager(diagnostics, null, null);
                Iterable<? extends JavaFileObject> compilationUnits = fileManager.getJavaFileObjects(sourceFile.toFile());

                JavaCompiler.CompilationTask task = compiler.getTask(
                    null,      // Writer for additional output
                    fileManager,
                    diagnostics,
                    Arrays.asList("-d", tempDir.toString()), // Output directory
                    null,      // Classes to process
                    compilationUnits
                );

                boolean success = task.call();

                fileManager.close();

                // Prepare response
                StringBuilder output = new StringBuilder();
                if (success) {
                    output.append("Compilation successful\n");

                    // List compiled files
                    Files.walk(tempDir)
                        .filter(path -> path.toString().endsWith(".class"))
                        .forEach(path -> output.append("Generated: ").append(path.getFileName()).append("\n"));
                } else {
                    output.append("Compilation failed\n");
                    for (Diagnostic<? extends JavaFileObject> diagnostic : diagnostics.getDiagnostics()) {
                        output.append(String.format("Error on line %d: %s\n",
                            diagnostic.getLineNumber(), diagnostic.getMessage(null)));
                    }
                }

                return objectMapper.writeValueAsString(Map.of(
                    "success", success,
                    "output", output.toString().trim(),
                    "errors", success ? new String[0] : diagnostics.getDiagnostics().stream()
                        .map(d -> d.getMessage(null))
                        .toArray(String[]::new)
                ));

            } finally {
                // Cleanup
                Files.walk(tempDir)
                    .sorted((a, b) -> b.compareTo(a))
                    .forEach(path -> {
                        try {
                            Files.delete(path);
                        } catch (IOException e) {
                            logger.warn("Failed to delete temp file: {}", path, e);
                        }
                    });
            }

        } catch (Exception e) {
            logger.error("Error in Java compilation", e);
            return objectMapper.writeValueAsString(Map.of(
                "success", false,
                "output", "",
                "errors", new String[]{e.getMessage()}
            ));
        }
    }
}