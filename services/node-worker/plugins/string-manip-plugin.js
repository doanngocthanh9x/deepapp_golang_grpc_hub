const BasePlugin = require('./base-plugin');

/**
 * String Manipulation Plugin
 */
class StringManipPlugin extends BasePlugin {
    getName() {
        return "string_ops";
    }

    getDescription() {
        return "Perform string operations (uppercase, lowercase, reverse, length)";
    }

    async execute(params, context) {
        const text = params.text || "";
        const operation = params.operation || "uppercase";

        let result;
        switch (operation) {
            case "uppercase":
                result = text.toUpperCase();
                break;
            case "lowercase":
                result = text.toLowerCase();
                break;
            case "reverse":
                result = text.split('').reverse().join('');
                break;
            case "length":
                result = text.length.toString();
                break;
            default:
                throw new Error(`Unknown operation: ${operation}`);
        }

        return {
            input: text,
            operation: operation,
            result: result,
            status: "success"
        };
    }
}

module.exports = StringManipPlugin;
