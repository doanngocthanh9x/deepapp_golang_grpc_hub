# DeepApp gRPC Hub - Python Worker SDK

This SDK provides a simple way to create workers that connect to the DeepApp gRPC Hub system using Python.

## Installation

```bash
pip install grpcio grpcio-tools
```

## Quick Start

```python
from deepapp_sdk import Worker

class MyWorker(Worker):
    def __init__(self):
        super().__init__(
            worker_id='my-python-worker',
            hub_address='localhost:50051'
        )

    def get_capabilities(self):
        return [
            {
                'name': 'hello',
                'description': 'Returns a hello message',
                'input_schema': '{}',
                'output_schema': '{"type":"object","properties":{"message":{"type":"string"}}}',
                'http_method': 'GET',
                'accepts_file': False
            },
            {
                'name': 'process_data',
                'description': 'Process uploaded data',
                'input_schema': '{"type":"object","properties":{"file":{"type":"string","format":"binary"}}}',
                'output_schema': '{"type":"object","properties":{"result":{"type":"string"}}}',
                'http_method': 'POST',
                'accepts_file': True,
                'file_field_name': 'file'
            }
        ]

    def handle_hello(self, message):
        """Handle hello capability"""
        return {
            'message': 'Hello from Python Worker! 🐍',
            'timestamp': datetime.now().isoformat(),
            'worker_id': self.worker_id
        }

    def handle_process_data(self, message):
        """Handle data processing capability"""
        content = json.loads(message.content)
        filename = content.get('filename', 'unknown')

        return {
            'filename': filename,
            'processed': True,
            'result': 'Data processed successfully',
            'timestamp': datetime.now().isoformat()
        }

# Start the worker
if __name__ == '__main__':
    worker = MyWorker()
    worker.start()
```

## API Reference

### Worker Class

#### Constructor

```python
Worker(worker_id='worker-id', hub_address='localhost:50051')
```

#### Methods

- `get_capabilities()` - Return list of capability definitions (override in subclass)
- `start()` - Connect to hub and start processing
- `stop()` - Disconnect from hub
- `handle_{capability_name}(message)` - Override to handle specific capabilities

#### Capability Definition

```python
{
    'name': 'capability_name',           # Unique capability name
    'description': 'What it does',       # Human readable description
    'input_schema': '{}',                # JSON Schema for input
    'output_schema': '{"type":"object"}', # JSON Schema for output
    'http_method': 'POST',               # HTTP method for web API
    'accepts_file': False,               # Whether it accepts file uploads
    'file_field_name': 'file'            # Field name for file uploads (if accepts_file=True)
}
```

#### Handler Methods

Implement handler methods for each capability:

```python
def handle_capability_name(self, message):
    # message.content contains the request data
    # Return response dict (will be JSON serialized)
    return {'result': 'success'}
```

## Advanced Usage

### File Upload Handling

```python
def handle_analyze_image(self, message):
    content = json.loads(message.content)

    # File data is base64 encoded in content['file']
    file_data = content['file']
    file_bytes = base64.b64decode(file_data)

    # Process the file...
    analysis = analyze_image(file_bytes)

    return {
        'analysis': analysis,
        'timestamp': datetime.now().isoformat()
    }
```

### Error Handling

```python
def handle_process_data(self, message):
    try:
        # Your processing logic
        return {'success': True}
    except Exception as e:
        return {
            'error': str(e),
            'status': 'failed'
        }
```

### Environment Variables

```bash
WORKER_ID=my-custom-worker
HUB_ADDRESS=localhost:50051
```

## Complete Example

See `examples/` directory for complete working examples.