package com.deepapp.worker.plugins;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Map;

/**
 * File Info Plugin - Reads file information
 */
public class FileInfoPlugin implements BasePlugin {
    
    private static final ObjectMapper objectMapper = new ObjectMapper();
    
    @Override
    public String getName() {
        return "read_file_info";
    }
    
    @Override
    public String getDescription() {
        return "Reads file information (size, modified time, etc.)";
    }
    
    @Override
    public String getInputSchema() {
        return "{\"type\":\"object\",\"properties\":{\"filePath\":{\"type\":\"string\"}}}";
    }
    
    @Override
    public String getOutputSchema() {
        return "{\"type\":\"object\",\"properties\":{\"filePath\":{\"type\":\"string\"},\"exists\":{\"type\":\"boolean\"},\"size\":{\"type\":\"number\"},\"lastModified\":{\"type\":\"string\"},\"isDirectory\":{\"type\":\"boolean\"}}}";
    }
    
    @Override
    public String execute(String input, Object workerSDK) throws Exception {
        Map<String, Object> params = objectMapper.readValue(input, Map.class);
        String filePath = (String) params.get("filePath");
        
        if (filePath == null || filePath.isEmpty()) {
            throw new IllegalArgumentException("filePath is required");
        }
        
        File file = new File(filePath);
        Map<String, Object> result = new HashMap<>();
        
        result.put("filePath", filePath);
        result.put("exists", file.exists());
        
        if (file.exists()) {
            result.put("size", file.length());
            result.put("lastModified", String.valueOf(file.lastModified()));
            result.put("isDirectory", file.isDirectory());
            result.put("canRead", file.canRead());
            result.put("canWrite", file.canWrite());
        } else {
            result.put("size", 0);
            result.put("lastModified", "");
            result.put("isDirectory", false);
        }
        
        return objectMapper.writeValueAsString(result);
    }
}
