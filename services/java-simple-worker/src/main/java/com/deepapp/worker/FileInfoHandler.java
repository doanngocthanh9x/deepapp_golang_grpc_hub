package com.deepapp.worker;

import com.deepapp.hub.Hub.ServiceCapability;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.attribute.BasicFileAttributes;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;

/**
 * File information capability handler
 */
public class FileInfoHandler implements CapabilityHandler {

    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public ServiceCapability getCapability() {
        return ServiceCapability.newBuilder()
                .setName("read_file_info")
                .setDescription("Reads file information (size, modified time, etc.)")
                .setInputSchema("{\"type\":\"object\",\"properties\":{\"filePath\":{\"type\":\"string\"}}}")
                .setOutputSchema("{\"type\":\"object\",\"properties\":{\"filePath\":{\"type\":\"string\"},\"exists\":{\"type\":\"boolean\"},\"size\":{\"type\":\"number\"},\"lastModified\":{\"type\":\"string\"},\"isDirectory\":{\"type\":\"boolean\"}}}")
                .build();
    }

    @Override
    public String handle(String input) throws Exception {
        // Parse input JSON
        Map<String, Object> request = objectMapper.readValue(input, Map.class);
        String filePath = (String) request.get("filePath");

        if (filePath == null || filePath.trim().isEmpty()) {
            throw new IllegalArgumentException("filePath is required");
        }

        File file = new File(filePath);
        Map<String, Object> response = new HashMap<>();

        response.put("filePath", filePath);
        response.put("exists", file.exists());

        if (file.exists()) {
            response.put("size", file.length());
            response.put("isDirectory", file.isDirectory());

            // Get last modified time
            long lastModified = file.lastModified();
            SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
            response.put("lastModified", sdf.format(new Date(lastModified)));

            // Try to get more detailed attributes if possible
            try {
                Path path = Paths.get(filePath);
                BasicFileAttributes attrs = Files.readAttributes(path, BasicFileAttributes.class);
                response.put("creationTime", sdf.format(new Date(attrs.creationTime().toMillis())));
                response.put("isRegularFile", attrs.isRegularFile());
                response.put("isSymbolicLink", attrs.isSymbolicLink());
            } catch (Exception e) {
                // Ignore if we can't get extended attributes
            }
        }

        return objectMapper.writeValueAsString(response);
    }
}