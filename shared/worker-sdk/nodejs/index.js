/**
 * Worker SDK for Node.js
 * Provides easy worker-to-worker communication and capability registration
 * 
 * Usage:
 *   const { WorkerSDK } = require('@deepapp/worker-sdk-nodejs');
 *   
 *   class MyWorker extends WorkerSDK {
 *     registerCapabilities() {
 *       this.addCapability('my_task', this.handleMyTask.bind(this), {
 *         description: 'Does something useful',
 *         httpMethod: 'POST'
 *       });
 *     }
 *     
 *     async handleMyTask(params) {
 *       // Call another worker if needed
 *       const result = await this.callWorker('other-worker', 'other_task', { key: 'value' });
 *       return { status: 'success', result };
 *     }
 *   }
 *   
 *   const worker = new MyWorker('my-worker', 'localhost:50051');
 *   worker.run();
 */

const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const { v4: uuidv4 } = require('uuid');
const path = require('path');
const EventEmitter = require('events');

class WorkerSDK extends EventEmitter {
  constructor(workerId, hubAddress, workerType = 'nodejs') {
    super();
    
    this.workerId = workerId;
    this.hubAddress = hubAddress;
    this.workerType = workerType;
    this.running = false;
    
    // Capability registry
    this.capabilities = {};
    this.capabilityHandlers = {};
    
    // Worker-to-worker call tracking
    this.pendingCalls = new Map();
    
    // Load proto files
    this.loadProto();
  }
  
  loadProto() {
    const PROTO_PATH = path.join(__dirname, '../../../proto/hub.proto');
    
    const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true
    });
    
    const protoDescriptor = grpc.loadPackageDefinition(packageDefinition);
    this.proto = protoDescriptor.hub;
    
    // Message types
    this.MessageType = {
      REGISTER: 1,
      REQUEST: 4,
      RESPONSE: 5,
      WORKER_CALL: 6
    };
  }
  
  /**
   * Override this method to register your worker's capabilities
   * Use this.addCapability() to register each capability
   */
  registerCapabilities() {
    throw new Error('registerCapabilities() must be implemented by subclass');
  }
  
  /**
   * Register a capability handler
   * 
   * @param {string} name - Capability name
   * @param {Function} handler - Async function that takes params and returns result
   * @param {Object} options - Capability options
   * @param {string} options.description - Human-readable description
   * @param {string} options.httpMethod - HTTP method (GET, POST, etc.)
   * @param {boolean} options.acceptsFile - Whether accepts file uploads
   * @param {string} options.fileFieldName - Name of file field
   * @param {string} options.inputSchema - JSON schema for input
   * @param {string} options.outputSchema - JSON schema for output
   */
  addCapability(name, handler, options = {}) {
    const {
      description = '',
      httpMethod = 'POST',
      acceptsFile = false,
      fileFieldName = '',
      inputSchema = '{}',
      outputSchema = '{}'
    } = options;
    
    this.capabilityHandlers[name] = handler;
    this.capabilities[name] = {
      name,
      description,
      input_schema: inputSchema,
      output_schema: outputSchema,
      http_method: httpMethod,
      accepts_file: acceptsFile
    };
    
    if (fileFieldName) {
      this.capabilities[name].file_field_name = fileFieldName;
    }
    
    this.log(`✓ Registered capability: ${name}`);
  }
  
  /**
   * Call another worker's capability through the Hub
   * 
   * @param {string} targetWorker - Worker ID to call
   * @param {string} capability - Capability name on target worker
   * @param {Object} params - Parameters to send
   * @param {number} timeout - Timeout in milliseconds (default 30000)
   * @returns {Promise<Object>} Response from target worker
   */
  callWorker(targetWorker, capability, params, timeout = 30000) {
    return new Promise((resolve, reject) => {
      if (!this.running) {
        return reject(new Error('Worker not connected. Call run() first'));
      }
      
      const requestId = uuidv4();
      
      this.log(`🔗 Calling ${targetWorker}.${capability}`);
      
      // Create worker call message
      const callMsg = {
        id: requestId,
        from: this.workerId,
        to: targetWorker,
        channel: capability,
        content: JSON.stringify(params),
        timestamp: new Date().toISOString(),
        type: this.MessageType.WORKER_CALL,
        metadata: { capability }
      };
      
      // Register pending call
      const timeoutId = setTimeout(() => {
        this.pendingCalls.delete(requestId);
        reject(new Error(`No response from ${targetWorker} after ${timeout}ms`));
      }, timeout);
      
      this.pendingCalls.set(requestId, { resolve, reject, timeoutId });
      
      // Send the call
      this.stream.write(callMsg);
    });
  }
  
  /**
   * Handle response from worker-to-worker call
   */
  handleWorkerCallResponse(msg) {
    const requestId = msg.metadata?.request_id;
    
    if (requestId && this.pendingCalls.has(requestId)) {
      const { resolve, reject, timeoutId } = this.pendingCalls.get(requestId);
      clearTimeout(timeoutId);
      this.pendingCalls.delete(requestId);
      
      try {
        const response = JSON.parse(msg.content);
        resolve(response);
      } catch (e) {
        reject(new Error(`Failed to parse response: ${e.message}`));
      }
    }
  }
  
  /**
   * Process incoming message
   */
  async processMessage(msg) {
    const channel = msg.channel;
    
    if (!this.capabilityHandlers[channel]) {
      return JSON.stringify({
        error: `Unknown capability: ${channel}`,
        status: 'failed'
      });
    }
    
    try {
      // Parse input
      const params = msg.content ? JSON.parse(msg.content) : {};
      
      // Call handler
      let result = await this.capabilityHandlers[channel](params);
      
      // Ensure result is object
      if (typeof result !== 'object') {
        result = { result };
      }
      
      return JSON.stringify(result);
      
    } catch (e) {
      this.log(`✗ Error in ${channel}: ${e.message}`);
      return JSON.stringify({
        error: e.message,
        status: 'failed'
      });
    }
  }
  
  /**
   * Send registration message to Hub
   */
  sendRegistration() {
    const capabilitiesList = Object.values(this.capabilities);
    
    const registrationData = {
      worker_id: this.workerId,
      worker_type: this.workerType,
      capabilities: capabilitiesList,
      metadata: {
        version: '1.0.0',
        sdk_version: '2.0.0'
      }
    };
    
    const registerMsg = {
      id: `register-${Date.now()}`,
      from: this.workerId,
      to: 'hub',
      channel: 'system',
      content: JSON.stringify(registrationData),
      timestamp: new Date().toISOString(),
      type: this.MessageType.REGISTER,
      metadata: {}
    };
    
    this.stream.write(registerMsg);
    this.log('📤 Sent registration');
  }
  
  /**
   * Handle incoming messages from Hub
   */
  async handleMessage(msg) {
    try {
      const msgType = msg.type;
      
      // Response from worker-to-worker call
      if (msgType === this.MessageType.RESPONSE) {
        this.handleWorkerCallResponse(msg);
        return;
      }
      
      // Another worker calling us or regular request
      const responseContent = await this.processMessage(msg);
      
      const responseMsg = {
        id: `resp-${Date.now()}`,
        from: this.workerId,
        to: msg.from,
        channel: msg.channel,
        content: responseContent,
        timestamp: new Date().toISOString(),
        type: this.MessageType.RESPONSE,
        metadata: {}
      };
      
      // Add request_id for worker-to-worker calls
      if (msgType === this.MessageType.WORKER_CALL) {
        responseMsg.metadata.request_id = msg.id;
        responseMsg.metadata.status = 'success';
      }
      
      this.stream.write(responseMsg);
      
    } catch (e) {
      this.log(`✗ Error processing message: ${e.message}`);
    }
  }
  
  /**
   * Start the worker and connect to Hub
   */
  run() {
    this.log('🚀 Starting Worker');
    this.log(`   ID: ${this.workerId}`);
    this.log(`   Hub: ${this.hubAddress}`);
    this.log('='.repeat(50));
    
    // Register capabilities
    this.registerCapabilities();
    this.log(`✓ Registered ${Object.keys(this.capabilities).length} capabilities`);
    
    try {
      // Connect to Hub
      this.log('Connecting to Hub...');
      const client = new this.proto.HubService(
        this.hubAddress,
        grpc.credentials.createInsecure()
      );
      
      // Start bidirectional stream
      this.stream = client.Connect();
      this.running = true;
      
      // Send registration
      this.sendRegistration();
      
      this.log('✓ Connected to Hub');
      this.log('📨 Listening for requests...\n');
      
      // Handle incoming messages
      this.stream.on('data', (msg) => {
        this.handleMessage(msg);
      });
      
      this.stream.on('end', () => {
        this.log('✗ Stream ended');
        this.running = false;
      });
      
      this.stream.on('error', (err) => {
        this.log(`✗ Stream error: ${err.message}`);
        this.running = false;
      });
      
      // Keep process alive
      process.on('SIGINT', () => {
        this.log('\n\n✗ Shutting down...');
        this.stop();
        process.exit(0);
      });
      
    } catch (e) {
      this.log(`✗ Error: ${e.message}`);
      this.stop();
    }
  }
  
  /**
   * Stop the worker
   */
  stop() {
    this.running = false;
    if (this.stream) {
      this.stream.end();
    }
    this.log('✗ Disconnected from Hub');
  }
  
  /**
   * Log a message
   */
  log(message) {
    console.log(`[${this.workerId}] ${message}`);
  }
}

module.exports = { WorkerSDK };
