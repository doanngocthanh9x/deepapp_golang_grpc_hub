const BasePlugin = require('./base-plugin');

/**
 * Hello Plugin - Simple greeting capability
 */
class HelloPlugin extends BasePlugin {
    getName() {
        return "hello_node";
    }

    getDescription() {
        return "Returns a hello message from Node.js worker";
    }

    getHttpMethod() {
        return "POST";
    }

    async execute(params, context) {
        const name = params.name || "World";
        
        return {
            message: `Hello ${name} from Node.js! 🟢`,
            worker_id: context.workerId,
            timestamp: new Date().toISOString(),
            node_version: process.version,
            status: "success"
        };
    }
}

module.exports = HelloPlugin;
