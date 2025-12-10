const BasePlugin = require('./base-plugin');

/**
 * JSON Processor Plugin
 */
class JsonProcessorPlugin extends BasePlugin {
    getName() {
        return "json_process";
    }

    getDescription() {
        return "Process JSON data (validate, format, extract keys)";
    }

    async execute(params, context) {
        const data = params.data;
        const action = params.action || "validate";

        let result;
        try {
            switch (action) {
                case "validate":
                    // Already parsed if we got here
                    result = { valid: true, message: "JSON is valid" };
                    break;
                    
                case "keys":
                    result = { keys: Object.keys(data || {}) };
                    break;
                    
                case "values":
                    result = { values: Object.values(data || {}) };
                    break;
                    
                case "pretty":
                    result = { formatted: JSON.stringify(data, null, 2) };
                    break;
                    
                default:
                    throw new Error(`Unknown action: ${action}`);
            }

            return {
                action: action,
                result: result,
                status: "success"
            };
            
        } catch (error) {
            return {
                action: action,
                error: error.message,
                status: "error"
            };
        }
    }
}

module.exports = JsonProcessorPlugin;
