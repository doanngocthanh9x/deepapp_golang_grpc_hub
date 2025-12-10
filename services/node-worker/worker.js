#!/usr/bin/env node
/**
 * Example Node.js Worker using the WorkerSDK
 * Demonstrates how easy it is to create a worker with worker-to-worker communication
 */

const { WorkerSDK } = require('../../../shared/worker-sdk/nodejs');

class NodeJSWorker extends WorkerSDK {
  registerCapabilities() {
    // Simple hello capability
    this.addCapability('hello_node', this.handleHello.bind(this), {
      description: 'Returns a hello message from Node.js',
      httpMethod: 'POST'
    });
    
    // Process data capability
    this.addCapability('process_data', this.handleProcessData.bind(this), {
      description: 'Process some data',
      httpMethod: 'POST',
      inputSchema: '{"type":"object","properties":{"data":{"type":"string"}}}',
      outputSchema: '{"type":"object","properties":{"processed":{"type":"string"}}}'
    });
    
    // Composite task - calls Python worker
    this.addCapability('node_composite', this.handleComposite.bind(this), {
      description: 'Calls Python worker for processing',
      httpMethod: 'POST'
    });
  }
  
  async handleHello(params) {
    return {
      message: 'Hello from Node.js Worker! 🟢',
      worker_id: this.workerId,
      status: 'success',
      timestamp: new Date().toISOString()
    };
  }
  
  async handleProcessData(params) {
    const data = params.data || 'no data';
    
    return {
      processed: data.toUpperCase(),
      length: data.length,
      worker_id: this.workerId,
      status: 'success'
    };
  }
  
  async handleComposite(params) {
    // Step 1: Do local processing
    const nodeResult = {
      processed_by: 'nodejs',
      timestamp: new Date().toISOString()
    };
    
    // Step 2: Call Python worker
    try {
      this.log('  → Calling Python worker...');
      const pythonResponse = await this.callWorker(
        'python-worker',
        'hello',
        {},
        30000
      );
      
      return {
        node_processing: nodeResult,
        python_response: pythonResponse,
        combined_status: 'success',
        worker_id: this.workerId
      };
      
    } catch (e) {
      // Return partial result on error
      return {
        node_processing: nodeResult,
        python_call_error: e.message,
        combined_status: 'partial',
        worker_id: this.workerId
      };
    }
  }
}

// Main
const workerId = process.env.WORKER_ID || 'nodejs-worker';
const hubAddress = process.env.HUB_ADDRESS || 'localhost:50051';

const worker = new NodeJSWorker(workerId, hubAddress);
worker.run();
