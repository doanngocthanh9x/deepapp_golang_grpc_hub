const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

/**
 * DeepApp gRPC Hub - Node.js Worker SDK
 *
 * This SDK provides a base Worker class for creating workers that connect
 * to the DeepApp gRPC Hub system.
 */

class Worker {
  constructor(options = {}) {
    this.workerId = options.workerId || `nodejs-worker-${Date.now()}`;
    this.hubAddress = options.hubAddress || 'localhost:50051';
    this.channel = null;
    this.client = null;
    this.stream = null;
    this.running = false;

    // Load protobuf definition
    this.loadProto();
  }

  loadProto() {
    const PROTO_PATH = __dirname + '/../../proto/hub.proto';

    const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true
    });

    this.proto = grpc.loadPackageDefinition(packageDefinition).hub;
  }

  getCapabilities() {
    // Override in subclass
    return [];
  }

  async start() {
    console.log(`🚀 Starting Node.js Worker: ${this.workerId}`);
    console.log(`📡 Connecting to Hub at: ${this.hubAddress}`);

    try {
      // Create gRPC client
      this.channel = new grpc.Channel(this.hubAddress, grpc.credentials.createInsecure());
      this.client = new this.proto.HubService(this.hubAddress, grpc.credentials.createInsecure());

      // Start bidirectional stream
      this.stream = this.client.Connect();

      this.running = true;

      // Handle incoming messages
      this.stream.on('data', (message) => {
        this.handleMessage(message);
      });

      this.stream.on('error', (error) => {
        console.error('Stream error:', error);
        this.running = false;
      });

      this.stream.on('end', () => {
        console.log('Stream ended');
        this.running = false;
      });

      // Send registration
      await this.sendRegistration();

      console.log(`✅ Worker ${this.workerId} connected and registered`);

      // Keep alive
      while (this.running) {
        await new Promise(resolve => setTimeout(resolve, 1000));
      }

    } catch (error) {
      console.error('Failed to start worker:', error);
      throw error;
    }
  }

  async sendRegistration() {
    const capabilities = this.getCapabilities();

    const registrationData = {
      worker_id: this.workerId,
      worker_type: 'nodejs',
      capabilities: capabilities,
      metadata: {
        version: '1.0.0',
        description: 'Node.js worker',
        sdk: 'nodejs-sdk'
      }
    };

    const message = {
      id: `reg-${Date.now()}`,
      from: this.workerId,
      to: 'hub',
      channel: 'system',
      content: JSON.stringify(registrationData),
      timestamp: new Date().toISOString(),
      type: 'REGISTER',
      action: 'register'
    };

    this.stream.write(message);
  }

  async handleMessage(message) {
    console.log(`📨 Received: ${message.channel} from ${message.from}`);

    try {
      // Route to capability handler
      const capability = message.channel;
      const handlerMethod = `handle${this.capitalize(capability)}`;

      if (typeof this[handlerMethod] === 'function') {
        const result = await this[handlerMethod](message);
        await this.sendResponse(message, result);
      } else {
        console.warn(`No handler for capability: ${capability}`);
        await this.sendResponse(message, {
          error: `Unknown capability: ${capability}`,
          status: 'failed'
        });
      }
    } catch (error) {
      console.error('Error handling message:', error);
      await this.sendResponse(message, {
        error: error.message,
        status: 'failed'
      });
    }
  }

  async sendResponse(requestMessage, responseData) {
    const responseMessage = {
      id: `resp-${Date.now()}`,
      from: this.workerId,
      to: requestMessage.from,
      channel: requestMessage.channel,
      content: JSON.stringify(responseData),
      timestamp: new Date().toISOString(),
      type: 'DIRECT',
      action: 'response'
    };

    this.stream.write(responseMessage);
    console.log(`📤 Sent response to ${requestMessage.from}`);
  }

  capitalize(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
  }

  stop() {
    console.log('🛑 Stopping worker...');
    this.running = false;

    if (this.stream) {
      this.stream.end();
    }

    if (this.channel) {
      this.channel.close();
    }
  }
}

module.exports = { Worker };