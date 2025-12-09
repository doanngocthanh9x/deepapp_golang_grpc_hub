package com.deepapp.worker;

import com.deepapp.hub.ServiceCapability;

/**
 * Interface for capability handlers
 */
public interface CapabilityHandler {

    /**
     * Get the capability definition
     */
    ServiceCapability getCapability();

    /**
     * Handle the capability execution
     */
    String handle(String input) throws Exception;
}