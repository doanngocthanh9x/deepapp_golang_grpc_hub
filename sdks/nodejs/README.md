# DeepApp gRPC Hub - Node.js Worker SDK

This SDK provides a simple way to create workers that connect to the DeepApp gRPC Hub system.

## Installation

```bash
npm install @grpc/grpc-js @grpc/proto-loader
```

## Quick Start

```javascript
const { Worker } = require('./worker-sdk');

class MyWorker extends Worker {
  constructor() {
    super({
      workerId: 'my-nodejs-worker',
      hubAddress: 'localhost:50051'
    });
  }

  // Define your capabilities
  getCapabilities() {
    return [
      {
        name: 'hello',
        description: 'Returns a hello message',
        inputSchema: '{}',
        outputSchema: '{"type":"object","properties":{"message":{"type":"string"}}}',
        httpMethod: 'GET',
        acceptsFile: false
      },
      {
        name: 'process_data',
        description: 'Process uploaded data',
        inputSchema: '{"type":"object","properties":{"file":{"type":"string","format":"binary"}}}',
        outputSchema: '{"type":"object","properties":{"result":{"type":"string"}}}',
        httpMethod: 'POST',
        acceptsFile: true,
        fileFieldName: 'file'
      }
    ];
  }

  // Implement capability handlers
  async handleHello(message) {
    return {
      message: 'Hello from Node.js Worker! 🚀',
      timestamp: new Date().toISOString(),
      workerId: this.workerId
    };
  }

  async handleProcessData(message) {
    const content = JSON.parse(message.content);
    const filename = content.filename || 'unknown';

    return {
      filename,
      processed: true,
      result: 'Data processed successfully',
      timestamp: new Date().toISOString()
    };
  }
}

// Start the worker
const worker = new MyWorker();
worker.start().catch(console.error);
```

## API Reference

### Worker Class

#### Constructor Options

```javascript
const options = {
  workerId: 'my-worker',        // Unique worker ID
  hubAddress: 'localhost:50051' // Hub gRPC address
};
```

#### Methods

- `getCapabilities()` - Return array of capability definitions
- `start()` - Connect to hub and start processing
- `stop()` - Disconnect from hub

#### Capability Definition

```javascript
{
  name: 'capability_name',           // Unique capability name
  description: 'What it does',       // Human readable description
  inputSchema: '{}',                 // JSON Schema for input
  outputSchema: '{"type":"object"}', // JSON Schema for output
  httpMethod: 'POST',                // HTTP method for web API
  acceptsFile: false,                // Whether it accepts file uploads
  fileFieldName: 'file'              // Field name for file uploads (if acceptsFile=true)
}
```

#### Handler Methods

Implement handler methods for each capability:

```javascript
async handleCapabilityName(message) {
  // message.content contains the request data
  // Return response object (will be JSON.stringify'd)
  return { result: 'success' };
}
```

## Advanced Usage

### File Upload Handling

```javascript
async handleAnalyzeImage(message) {
  const content = JSON.parse(message.content);

  // File data is base64 encoded in content.file
  const fileBuffer = Buffer.from(content.file, 'base64');

  // Process the file...
  const analysis = await analyzeImage(fileBuffer);

  return {
    analysis,
    timestamp: new Date().toISOString()
  };
}
```

### Error Handling

```javascript
async handleCapability(message) {
  try {
    // Your logic here
    return { success: true };
  } catch (error) {
    return {
      error: error.message,
      status: 'failed'
    };
  }
}
```

### Environment Variables

```bash
WORKER_ID=my-custom-worker
HUB_ADDRESS=localhost:50051
```

## Complete Example

See `examples/` directory for complete working examples.</content>
<parameter name="filePath">/home/vps1/WorkSpace/deepapp_golang_grpc_hub/sdks/nodejs/README.md