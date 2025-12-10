package com.deepapp.worker.plugins;

import java.util.Map;

/**
 * Base interface for all worker plugins
 * 
 * Each plugin represents one capability that the worker can perform.
 * Simply create a new class in the plugins package that implements this interface.
 */
public interface BasePlugin {
    
    /**
     * Get the capability name (e.g., "hello_world", "process_data")
     */
    String getName();
    
    /**
     * Get human-readable description
     */
    String getDescription();
    
    /**
     * Get JSON schema for input validation (optional)
     */
    default String getInputSchema() {
        return "{}";
    }
    
    /**
     * Get JSON schema for output format (optional)
     */
    default String getOutputSchema() {
        return "{}";
    }
    
    /**
     * Get HTTP method for Web API endpoint
     */
    default String getHttpMethod() {
        return "POST";
    }
    
    /**
     * Whether this capability accepts file upload
     */
    default boolean acceptsFile() {
        return false;
    }
    
    /**
     * Execute the plugin logic
     * 
     * @param input Input parameters as JSON string
     * @param workerSDK Reference to worker SDK for calling other workers
     * @return Result as JSON string
     * @throws Exception if execution fails
     */
    String execute(String input, Object workerSDK) throws Exception;
    
    /**
     * Called when plugin is loaded (optional hook)
     */
    default void onLoad() {
        // Override if needed
    }
    
    /**
     * Called when plugin is unloaded (optional hook)
     */
    default void onUnload() {
        // Override if needed
    }
}
