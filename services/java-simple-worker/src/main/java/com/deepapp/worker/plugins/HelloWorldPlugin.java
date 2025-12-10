package com.deepapp.worker.plugins;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.util.HashMap;
import java.util.Map;

/**
 * Hello World Plugin - Simple greeting capability
 */
public class HelloWorldPlugin implements BasePlugin {
    
    private static final ObjectMapper objectMapper = new ObjectMapper();
    
    @Override
    public String getName() {
        return "hello_world";
    }
    
    @Override
    public String getDescription() {
        return "Returns a hello world message";
    }
    
    @Override
    public String getOutputSchema() {
        return "{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}";
    }
    
    @Override
    public String execute(String input, Object workerSDK) throws Exception {
        Map<String, Object> result = new HashMap<>();
        result.put("message", "Hello from Java Worker! ☕");
        result.put("timestamp", System.currentTimeMillis());
        result.put("status", "success");
        
        return objectMapper.writeValueAsString(result);
    }
}
