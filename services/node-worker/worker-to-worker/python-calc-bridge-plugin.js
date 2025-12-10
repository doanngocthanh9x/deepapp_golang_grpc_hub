const BasePlugin = require('../plugins/base-plugin');

/**
 * Python Calculator Bridge Plugin
 * Calls Python worker's calculator capability
 */
class PythonCalcBridgePlugin extends BasePlugin {
    getName() {
        return "python_calc_bridge";
    }

    getDescription() {
        return "Bridge to Python worker calculator - performs math via Python";
    }

    async execute(params, context) {
        const { operation, a, b } = params;
        
        if (!operation || a === undefined || b === undefined) {
            throw new Error("Missing required parameters: operation, a, b");
        }

        console.log(`🔄 Calling Python calculator: ${a} ${operation} ${b}`);
        
        try {
            const pythonResponse = await context.callWorker(
                'python-worker',
                'calculate',
                { operation, a, b },
                10000
            );
            
            return {
                bridge_from: 'node-worker',
                calculated_by: 'python-worker',
                input: { operation, a, b },
                python_response: pythonResponse,
                status: 'success'
            };
            
        } catch (error) {
            return {
                bridge_from: 'node-worker',
                calculated_by: 'python-worker',
                input: { operation, a, b },
                error: error.message,
                status: 'error'
            };
        }
    }
}

module.exports = PythonCalcBridgePlugin;
