/**
 * Base Plugin Class for Node.js Worker
 * All plugins should extend this class
 */

class BasePlugin {
    constructor() {
        if (this.constructor === BasePlugin) {
            throw new Error("BasePlugin is abstract and cannot be instantiated directly");
        }
    }

    /**
     * Get plugin name
     * @returns {string}
     */
    getName() {
        throw new Error("getName() must be implemented");
    }

    /**
     * Get plugin description
     * @returns {string}
     */
    getDescription() {
        return "";
    }

    /**
     * Get HTTP method (GET, POST, etc.)
     * @returns {string}
     */
    getHttpMethod() {
        return "POST";
    }

    /**
     * Does this plugin accept file uploads?
     * @returns {boolean}
     */
    acceptsFile() {
        return false;
    }

    /**
     * Get file field name for uploads
     * @returns {string}
     */
    getFileFieldName() {
        return "file";
    }

    /**
     * Execute the plugin logic
     * @param {Object} params - Request parameters
     * @param {Object} context - Execution context (worker instance, etc.)
     * @returns {Promise<Object>}
     */
    async execute(params, context) {
        throw new Error("execute() must be implemented");
    }
}

module.exports = BasePlugin;
