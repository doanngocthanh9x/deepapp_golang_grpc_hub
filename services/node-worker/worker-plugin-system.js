#!/usr/bin/env node

const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');
const fs = require('fs');

const PluginManager = require('./plugins/plugin-manager');

// Load proto
const PROTO_PATH = path.join(__dirname, '../proto/hub.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});
const hubProto = grpc.loadPackageDefinition(packageDefinition).hub;

/**
 * Node.js Worker with Plugin System
 */
class NodeWorker {
    constructor(workerId, hubAddress) {
        this.workerId = workerId;
        this.hubAddress = hubAddress;
        this.pluginManager = new PluginManager();
        this.stream = null;
        this.pendingCalls = new Map();
        this.connected = false;
    }

    /**
     * Initialize worker
     */
    async initialize() {
        console.log('🟢 Node.js Worker Starting...');
        console.log(`   Worker ID: ${this.workerId}`);
        console.log(`   Hub Address: ${this.hubAddress}`);
        
        // Discover and load all plugins
        console.log('\n📦 Loading plugins...');
        this.pluginManager.discoverPlugins();
        
        // Also load worker-to-worker plugins
        this.loadWorkerToWorkerPlugins();
        
        const capabilities = this.pluginManager.getCapabilities();
        console.log(`\n✅ Registered ${capabilities.length} capabilities:`);
        capabilities.forEach(cap => {
            console.log(`   - ${cap.name}: ${cap.description}`);
        });
    }

    /**
     * Load worker-to-worker communication plugins
     */
    loadWorkerToWorkerPlugins() {
        const workerToWorkerDir = path.join(__dirname, 'worker-to-worker');
        
        if (!fs.existsSync(workerToWorkerDir)) {
            return;
        }

        console.log(`\n🔄 Loading worker-to-worker plugins...`);
        const files = fs.readdirSync(workerToWorkerDir);
        
        for (const file of files) {
            if (file.endsWith('-plugin.js')) {
                try {
                    const PluginClass = require(path.join(workerToWorkerDir, file));
                    const plugin = new PluginClass();
                    const pluginName = plugin.getName();
                    
                    this.pluginManager.plugins.set(pluginName, plugin);
                    console.log(`   ✓ Loaded: ${pluginName} (${file})`);
                } catch (error) {
                    console.error(`   ✗ Failed to load ${file}:`, error.message);
                }
            }
        }
    }

    /**
     * Connect to Hub and establish bidirectional stream
     */
    connect() {
        const client = new hubProto.HubService(
            this.hubAddress,
            grpc.credentials.createInsecure()
        );

        this.stream = client.Connect();

        // Handle incoming messages from Hub
        this.stream.on('data', (message) => {
            this.handleHubMessage(message);
        });

        this.stream.on('error', (error) => {
            console.error('❌ Stream error:', error.message);
            this.connected = false;
            setTimeout(() => this.reconnect(), 5000);
        });

        this.stream.on('end', () => {
            console.log('⚠️  Stream ended');
            this.connected = false;
            setTimeout(() => this.reconnect(), 5000);
        });

        // Register worker with capabilities
        this.register();
    }

    /**
     * Register worker with Hub
     */
    register() {
        const capabilities = this.pluginManager.getCapabilities();
        
        const registrationData = {
            worker_id: this.workerId,
            worker_type: 'node',
            capabilities: capabilities,
            metadata: {
                version: '1.0.0',
                description: 'Node.js worker with plugin system',
                plugin_count: capabilities.length
            }
        };

        const registerMessage = {
            type: 3, // REGISTER = 3
            from: this.workerId,
            to: 'hub',
            content: JSON.stringify(registrationData),
            action: 'register'
        };

        this.stream.write(registerMessage);
        this.connected = true;
        console.log('\n✅ Worker registered with Hub');
        
        // Send heartbeat every 30 seconds
        this.startHeartbeat();
    }

    /**
     * Start heartbeat to keep connection alive
     */
    startHeartbeat() {
        setInterval(() => {
            if (this.connected && this.stream) {
                this.stream.write({
                    type: 'HEARTBEAT',
                    from: this.workerId,
                    heartbeat: { timestamp: new Date().toISOString() }
                });
            }
        }, 30000);
    }

    /**
     * Handle incoming messages from Hub
     */
    async handleHubMessage(message) {
        // MessageType enum: DIRECT=0, BROADCAST=1, CHANNEL=2, REGISTER=3, REQUEST=4, RESPONSE=5, WORKER_CALL=6
        // Note: gRPC may send as number OR string depending on serialization
        
        try {
            // Normalize type to string for comparison
            const messageType = typeof message.type === 'string' ? message.type : String(message.type);
            
            switch (messageType) {
                case '4':
                case 'REQUEST':
                    await this.handleRequest(message);
                    break;
                    
                case '5':
                case 'RESPONSE':
                    this.handleResponse(message);
                    break;
                    
                case '3':
                case 'REGISTER':
                    console.log('✓ Registration acknowledged');
                    break;
                    
                default:
                    // Silently ignore heartbeats and unknown types
                    break;
            }
        } catch (error) {
            console.error('❌ Error handling message:', error.message);
        }
    }

    /**
     * Handle capability request
     */
    async handleRequest(message) {
        // Extract request info from message fields (not content)
        const request_id = message.id;  // Request ID is in message.id
        const original_sender = message.from;  // Who sent the request (web-api)
        const capability = message.metadata?.capability || message.channel;  // Capability in metadata or channel
        
        // Parse user data from message.content
        let params = {};
        if (message.content && message.content.trim() !== '' && message.content !== '{}') {
            try {
                params = JSON.parse(message.content);
            } catch (e) {
                console.error('❌ Failed to parse content:', e.message);
                return;
            }
        }
        
        console.log(`📥 Request: ${capability} (ID: ${request_id})`);
        
        try {
            // Get plugin
            const plugin = this.pluginManager.getPlugin(capability);
            if (!plugin) {
                throw new Error(`Unknown capability: ${capability}`);
            }

            // Execute plugin with context
            const context = {
                workerId: this.workerId,
                callWorker: this.callWorker.bind(this)
            };
            
            const result = await plugin.execute(params, context);

            // Send response back to original sender
            this.sendResponse(request_id, original_sender, result, 'success');
            console.log(`✅ Response sent for: ${capability}`);
            
        } catch (error) {
            console.error(`❌ Error processing ${capability}:`, error.message);
            this.sendResponse(request_id, original_sender, { error: error.message }, 'error');
        }
    }

    /**
     * Send response to Hub
     */
    sendResponse(requestId, targetClient, result, status) {
        // Send response back - the result data goes directly in content as JSON
        const responseMessage = {
            id: requestId,  // Same ID as request
            type: 5, // RESPONSE = 5
            from: this.workerId,
            to: targetClient,  // Send back to original requester (web-api)
            content: JSON.stringify(result),  // Just the result data
            timestamp: new Date().toISOString()
        };

        this.stream.write(responseMessage);
    }

    /**
     * Call another worker's capability
     */
    callWorker(targetWorkerId, capability, params, timeout = 30000) {
        return new Promise((resolve, reject) => {
            const requestId = `${this.workerId}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
            
            // Store pending call
            const timer = setTimeout(() => {
                this.pendingCalls.delete(requestId);
                reject(new Error(`Timeout calling ${targetWorkerId}.${capability}`));
            }, timeout);

            this.pendingCalls.set(requestId, {
                resolve,
                reject,
                timer,
                targetWorkerId,
                capability
            });

            // Send WORKER_CALL message to Hub
            const callMessage = {
                type: 'WORKER_CALL',
                from: this.workerId,
                worker_call: {
                    request_id: requestId,
                    target_worker_id: targetWorkerId,
                    capability: capability,
                    data: JSON.stringify(params)
                }
            };

            this.stream.write(callMessage);
        });
    }

    /**
     * Handle response from another worker
     */
    handleResponse(message) {
        const { request_id, status, data } = message.response;
        
        const pending = this.pendingCalls.get(request_id);
        if (!pending) {
            return;
        }

        clearTimeout(pending.timer);
        this.pendingCalls.delete(request_id);

        try {
            const result = JSON.parse(data);
            
            if (status === 'success') {
                pending.resolve(result);
            } else {
                pending.reject(new Error(result.error || 'Worker call failed'));
            }
        } catch (error) {
            pending.reject(error);
        }
    }

    /**
     * Reconnect to Hub
     */
    reconnect() {
        console.log('🔄 Reconnecting to Hub...');
        this.connect();
    }

    /**
     * Start the worker
     */
    async start() {
        await this.initialize();
        this.connect();
        
        console.log('\n🚀 Node.js Worker is running!\n');
    }
}

// Main execution
const workerId = process.env.WORKER_ID || 'node-worker';
const hubAddress = process.env.HUB_ADDRESS || 'localhost:50051';

const worker = new NodeWorker(workerId, hubAddress);
worker.start().catch(error => {
    console.error('❌ Fatal error:', error);
    process.exit(1);
});

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('\n👋 Shutting down gracefully...');
    process.exit(0);
});

process.on('SIGTERM', () => {
    console.log('\n👋 Shutting down gracefully...');
    process.exit(0);
});
