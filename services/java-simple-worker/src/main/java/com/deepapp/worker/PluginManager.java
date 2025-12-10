package com.deepapp.worker;

import com.deepapp.worker.plugins.BasePlugin;
import org.reflections.Reflections;
import org.reflections.scanners.Scanners;
import org.reflections.util.ClasspathHelper;
import org.reflections.util.ConfigurationBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.*;

/**
 * Plugin Manager - Auto-discovers and loads all plugins
 * 
 * Scans the plugins package for classes implementing BasePlugin
 * and automatically loads them.
 */
public class PluginManager {
    
    private static final Logger logger = LoggerFactory.getLogger(PluginManager.class);
    
    private final Map<String, BasePlugin> plugins = new HashMap<>();
    private final String pluginPackage;
    
    public PluginManager() {
        this("com.deepapp.worker.plugins");
    }
    
    public PluginManager(String pluginPackage) {
        this.pluginPackage = pluginPackage;
    }
    
    /**
     * Auto-discover and load all plugins from the plugins package
     */
    public void loadAllPlugins() {
        logger.info("🔌 Auto-discovering plugins from package: {}", pluginPackage);
        
        try {
            Reflections reflections = new Reflections(pluginPackage);
            Set<Class<? extends BasePlugin>> pluginClasses = 
                reflections.getSubTypesOf(BasePlugin.class);
            
            logger.info("📦 Found {} plugin classes", pluginClasses.size());
            
            for (Class<? extends BasePlugin> pluginClass : pluginClasses) {
                try {
                    // Skip interfaces and abstract classes
                    if (pluginClass.isInterface() || 
                        java.lang.reflect.Modifier.isAbstract(pluginClass.getModifiers())) {
                        continue;
                    }
                    
                    // Instantiate plugin
                    BasePlugin plugin = pluginClass.getDeclaredConstructor().newInstance();
                    String capabilityName = plugin.getName();
                    
                    // Register plugin
                    plugins.put(capabilityName, plugin);
                    
                    // Call on_load hook
                    plugin.onLoad();
                    
                    logger.info("  ✓ Loaded plugin: {} → capability '{}'", 
                        pluginClass.getSimpleName(), capabilityName);
                    
                } catch (Exception e) {
                    logger.error("  ✗ Error loading plugin {}: {}", 
                        pluginClass.getSimpleName(), e.getMessage());
                }
            }
            
            logger.info("✅ Successfully loaded {} plugins\n", plugins.size());
            
            // Also load worker-to-worker plugins
            loadWorkerToWorkerPlugins();
            
        } catch (Exception e) {
            logger.error("Error discovering plugins: {}", e.getMessage(), e);
        }
    }
    
    /**
     * Load plugins from worker-to-worker package
     */
    private void loadWorkerToWorkerPlugins() {
        try {
            logger.info("🔄 Loading worker-to-worker plugins...");
            
            Reflections workerToWorkerReflections = new Reflections(
                new ConfigurationBuilder()
                    .setUrls(ClasspathHelper.forPackage("com.deepapp.worker.workertoworker"))
                    .setScanners(Scanners.SubTypes)
            );
            
            Set<Class<? extends BasePlugin>> workerToWorkerPlugins = 
                workerToWorkerReflections.getSubTypesOf(BasePlugin.class);
            
            for (Class<? extends BasePlugin> pluginClass : workerToWorkerPlugins) {
                try {
                    if (pluginClass.isInterface() || 
                        java.lang.reflect.Modifier.isAbstract(pluginClass.getModifiers())) {
                        continue;
                    }
                    
                    BasePlugin plugin = pluginClass.getDeclaredConstructor().newInstance();
                    String capabilityName = plugin.getName();
                    plugins.put(capabilityName, plugin);
                    plugin.onLoad();
                    
                    logger.info("  ✓ Loaded: {} → '{}'", 
                        pluginClass.getSimpleName(), capabilityName);
                        
                } catch (Exception e) {
                    logger.error("  ✗ Error loading {}: {}", 
                        pluginClass.getSimpleName(), e.getMessage());
                }
            }
            
            if (workerToWorkerPlugins.size() > 0) {
                logger.info("✅ Loaded {} worker-to-worker plugins\n", workerToWorkerPlugins.size());
            }
            
        } catch (Exception e) {
            logger.error("Error loading worker-to-worker plugins: {}", e.getMessage());
        }
    }
    
    /**
     * Get a plugin by capability name
     */
    public BasePlugin getPlugin(String capabilityName) {
        return plugins.get(capabilityName);
    }
    
    /**
     * Get all loaded plugins
     */
    public Map<String, BasePlugin> getAllPlugins() {
        return new HashMap<>(plugins);
    }
    
    /**
     * Get all capabilities for registration with Hub
     */
    public List<Map<String, Object>> getAllCapabilities() {
        List<Map<String, Object>> capabilities = new ArrayList<>();
        
        for (BasePlugin plugin : plugins.values()) {
            Map<String, Object> capability = new HashMap<>();
            capability.put("name", plugin.getName());
            capability.put("description", plugin.getDescription());
            capability.put("input_schema", plugin.getInputSchema());
            capability.put("output_schema", plugin.getOutputSchema());
            capability.put("http_method", plugin.getHttpMethod());
            capability.put("accepts_file", plugin.acceptsFile());
            
            capabilities.add(capability);
        }
        
        return capabilities;
    }
    
    /**
     * Execute a plugin by capability name
     */
    public String executePlugin(String capabilityName, String input, Object workerSDK) throws Exception {
        BasePlugin plugin = getPlugin(capabilityName);
        
        if (plugin == null) {
            throw new IllegalArgumentException("Unknown capability: " + capabilityName);
        }
        
        return plugin.execute(input, workerSDK);
    }
    
    /**
     * Unload all plugins
     */
    public void unloadAllPlugins() {
        for (BasePlugin plugin : plugins.values()) {
            try {
                plugin.onUnload();
            } catch (Exception e) {
                logger.error("Error unloading plugin {}: {}", plugin.getName(), e.getMessage());
            }
        }
        plugins.clear();
    }
    
    /**
     * Get number of loaded plugins
     */
    public int getPluginCount() {
        return plugins.size();
    }
}
