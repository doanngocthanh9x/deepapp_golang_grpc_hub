const fs = require('fs');
const path = require('path');
const BasePlugin = require('./base-plugin');

/**
 * Plugin Manager - Auto-discovers and loads plugins
 */
class PluginManager {
    constructor() {
        this.plugins = new Map();
        this.pluginDir = __dirname;
    }

    /**
     * Discover and load all plugins from plugins directory
     */
    discoverPlugins() {
        console.log(`🔍 Discovering plugins in: ${this.pluginDir}`);
        
        const files = fs.readdirSync(this.pluginDir);
        
        for (const file of files) {
            // Skip base-plugin.js and plugin-manager.js
            if (file === 'base-plugin.js' || file === 'plugin-manager.js') {
                continue;
            }
            
            // Only load .js files ending with -plugin.js
            if (file.endsWith('-plugin.js')) {
                try {
                    this.loadPlugin(file);
                } catch (error) {
                    console.error(`❌ Failed to load plugin ${file}:`, error.message);
                }
            }
        }
        
        console.log(`✅ Loaded ${this.plugins.size} plugins`);
        return Array.from(this.plugins.values());
    }

    /**
     * Load a single plugin file
     */
    loadPlugin(filename) {
        const pluginPath = path.join(this.pluginDir, filename);
        
        // Require the plugin module
        const PluginClass = require(pluginPath);
        
        // Verify it extends BasePlugin
        if (!(PluginClass.prototype instanceof BasePlugin)) {
            throw new Error(`${filename} does not extend BasePlugin`);
        }
        
        // Instantiate the plugin
        const plugin = new PluginClass();
        const pluginName = plugin.getName();
        
        // Register the plugin
        this.plugins.set(pluginName, plugin);
        console.log(`  ✓ Loaded plugin: ${pluginName} (${filename})`);
    }

    /**
     * Get a plugin by name
     */
    getPlugin(name) {
        return this.plugins.get(name);
    }

    /**
     * Get all plugins
     */
    getAllPlugins() {
        return Array.from(this.plugins.values());
    }

    /**
     * Get plugin metadata for registration with Hub
     */
    getCapabilities() {
        const capabilities = [];
        
        for (const plugin of this.plugins.values()) {
            capabilities.push({
                name: plugin.getName(),
                description: plugin.getDescription(),
                http_method: plugin.getHttpMethod(),
                accepts_file: plugin.acceptsFile(),
                file_field_name: plugin.getFileFieldName()
            });
        }
        
        return capabilities;
    }
}

module.exports = PluginManager;
