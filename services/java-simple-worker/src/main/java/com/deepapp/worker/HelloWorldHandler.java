package com.deepapp.worker;

import com.deepapp.hub.Hub.ServiceCapability;

/**
 * Hello World capability handler
 */
public class HelloWorldHandler implements CapabilityHandler {

    @Override
    public ServiceCapability getCapability() {
        return ServiceCapability.newBuilder()
                .setName("hello_world")
                .setDescription("Returns a hello world message")
                .setInputSchema("{}")
                .setOutputSchema("{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}")
                .build();
    }

    @Override
    public String handle(String input) throws Exception {
        return "{\"message\":\"Hello World from Java Simple Worker!\"}";
    }
}