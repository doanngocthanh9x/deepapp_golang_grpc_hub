const BasePlugin = require('../plugins/base-plugin');

/**
 * Composite Plugin - Demonstrates worker-to-worker communication
 * Calls other workers to perform complex tasks
 */
class CompositePlugin extends BasePlugin {
    getName() {
        return "node_composite";
    }

    getDescription() {
        return "Calls multiple workers (Python, Java) and combines their results";
    }

    async execute(params, context) {
        console.log("🔄 Starting composite task...");
        
        const results = {
            timestamp: new Date().toISOString(),
            worker_id: context.workerId,
            calls: []
        };

        // Call Python worker
        try {
            console.log("  → Calling Python worker (hello)...");
            const pythonResponse = await context.callWorker(
                'python-worker',
                'hello',
                { name: "from Node.js" },
                10000
            );
            
            results.calls.push({
                worker: 'python-worker',
                capability: 'hello',
                status: 'success',
                response: pythonResponse
            });
            
        } catch (error) {
            console.error("  ✗ Python worker call failed:", error.message);
            results.calls.push({
                worker: 'python-worker',
                capability: 'hello',
                status: 'error',
                error: error.message
            });
        }

        // Call Java worker
        try {
            console.log("  → Calling Java worker (hello_world)...");
            const javaResponse = await context.callWorker(
                'java-simple-worker',
                'hello_world',
                {},
                10000
            );
            
            results.calls.push({
                worker: 'java-simple-worker',
                capability: 'hello_world',
                status: 'success',
                response: javaResponse
            });
            
        } catch (error) {
            console.error("  ✗ Java worker call failed:", error.message);
            results.calls.push({
                worker: 'java-simple-worker',
                capability: 'hello_world',
                status: 'error',
                error: error.message
            });
        }

        // Summary
        const successCount = results.calls.filter(c => c.status === 'success').length;
        results.summary = {
            total_calls: results.calls.length,
            successful: successCount,
            failed: results.calls.length - successCount
        };

        console.log("✅ Composite task completed");
        return results;
    }
}

module.exports = CompositePlugin;
