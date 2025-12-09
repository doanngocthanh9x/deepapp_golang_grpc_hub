package com.deepapp.sdk;

/**
 * Represents a capability that a worker can provide
 */
public class Capability {
    private final String name;
    private final String description;
    private final String inputSchema;
    private final String outputSchema;
    private final String httpMethod;
    private final boolean acceptsFile;
    private final String fileFieldName;

    public Capability(String name, String description, String inputSchema,
                     String outputSchema, String httpMethod, boolean acceptsFile,
                     String fileFieldName) {
        this.name = name;
        this.description = description;
        this.inputSchema = inputSchema;
        this.outputSchema = outputSchema;
        this.httpMethod = httpMethod;
        this.acceptsFile = acceptsFile;
        this.fileFieldName = fileFieldName;
    }

    // Getters
    public String getName() { return name; }
    public String getDescription() { return description; }
    public String getInputSchema() { return inputSchema; }
    public String getOutputSchema() { return outputSchema; }
    public String getHttpMethod() { return httpMethod; }
    public boolean isAcceptsFile() { return acceptsFile; }
    public String getFileFieldName() { return fileFieldName; }
}