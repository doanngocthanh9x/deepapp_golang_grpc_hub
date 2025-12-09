const { Worker } = require('../worker-sdk');

class ExampleWorker extends Worker {
  constructor() {
    super({
      workerId: 'nodejs-example-worker',
      hubAddress: process.env.HUB_ADDRESS || 'localhost:50051'
    });
  }

  // Define capabilities
  getCapabilities() {
    return [
      {
        name: 'hello',
        description: 'Returns a hello message',
        inputSchema: '{}',
        outputSchema: '{"type":"object","properties":{"message":{"type":"string"},"timestamp":{"type":"string"},"workerId":{"type":"string"}}}',
        httpMethod: 'GET',
        acceptsFile: false
      },
      {
        name: 'echo',
        description: 'Echoes back the input data',
        inputSchema: '{"type":"object","properties":{"message":{"type":"string"}}}',
        outputSchema: '{"type":"object","properties":{"echo":{"type":"string"},"timestamp":{"type":"string"}}}',
        httpMethod: 'POST',
        acceptsFile: false
      },
      {
        name: 'process_file',
        description: 'Process an uploaded file',
        inputSchema: '{"type":"object","properties":{"file":{"type":"string","format":"binary"},"filename":{"type":"string"}}}',
        outputSchema: '{"type":"object","properties":{"filename":{"type":"string"},"size":{"type":"number"},"processed":{"type":"boolean"},"timestamp":{"type":"string"}}}',
        httpMethod: 'POST',
        acceptsFile: true,
        fileFieldName: 'file'
      }
    ];
  }

  // Handler for hello capability
  async handleHello(message) {
    console.log('🔍 Processing hello request');

    return {
      message: 'Hello World from Node.js Worker! 🚀',
      timestamp: new Date().toISOString(),
      workerId: this.workerId,
      status: 'success'
    };
  }

  // Handler for echo capability
  async handleEcho(message) {
    console.log('🔍 Processing echo request');

    try {
      const content = JSON.parse(message.content);
      const inputMessage = content.message || 'No message provided';

      return {
        echo: inputMessage,
        timestamp: new Date().toISOString(),
        status: 'success'
      };
    } catch (error) {
      return {
        error: 'Invalid JSON input',
        status: 'failed'
      };
    }
  }

  // Handler for file processing capability
  async handleProcessFile(message) {
    console.log('🔍 Processing file upload');

    try {
      const content = JSON.parse(message.content);
      const filename = content.filename || 'unknown';
      const fileData = content.file;

      if (!fileData) {
        return {
          error: 'No file data provided',
          status: 'failed'
        };
      }

      // Decode base64 file data
      const fileBuffer = Buffer.from(fileData, 'base64');
      const fileSize = fileBuffer.length;

      // Simulate file processing
      console.log(`📁 Processing file: ${filename} (${fileSize} bytes)`);

      // Here you would do actual file processing
      // For example: image analysis, text extraction, etc.

      return {
        filename,
        size: fileSize,
        processed: true,
        result: 'File processed successfully',
        timestamp: new Date().toISOString(),
        status: 'success'
      };
    } catch (error) {
      console.error('Error processing file:', error);
      return {
        error: error.message,
        status: 'failed'
      };
    }
  }
}

// Start the worker
async function main() {
  const worker = new ExampleWorker();

  // Handle graceful shutdown
  process.on('SIGINT', () => {
    console.log('\n🛑 Received SIGINT, shutting down...');
    worker.stop();
    process.exit(0);
  });

  process.on('SIGTERM', () => {
    console.log('\n🛑 Received SIGTERM, shutting down...');
    worker.stop();
    process.exit(0);
  });

  try {
    await worker.start();
  } catch (error) {
    console.error('Failed to start worker:', error);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = ExampleWorker;