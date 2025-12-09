# DeepApp gRPC Hub - SDK Documentation

This directory contains SDK documentation and examples for connecting to the DeepApp gRPC Hub.

## Available SDKs

### 1. [Go SDK](./go/README.md)
- Full-featured Go client library
- Native gRPC support
- Examples for broadcast, direct, and channel messaging
- Reconnection and error handling utilities

### 2. [Python SDK](./python/README.md)
- Python client library with both sync and async support
- Easy-to-use API
- Context manager support
- Threading and asyncio examples

## Quick Links

- [Go SDK Documentation](./go/README.md)
- [Python SDK Documentation](./python/README.md)

## Protocol Buffer Definition

All SDKs use the same protocol buffer definition located at `/proto/hub.proto`:

```protobuf
syntax = "proto3";

package hub;

service HubService {
  rpc Connect(stream Message) returns (stream Message);
}

message Message {
  string id = 1;
  string from = 2;
  string to = 3;
  string channel = 4;
  string content = 5;
  string timestamp = 6;
  MessageType type = 7;
}

enum MessageType {
  DIRECT = 0;
  BROADCAST = 1;
  CHANNEL = 2;
}
```

## Message Types

### Broadcast Messages
- Sent to all connected clients
- No specific recipient
- Type: `BROADCAST`

### Direct Messages
- Sent to a specific client by ID
- Requires `to` field
- Type: `DIRECT`

### Channel Messages
- Sent to all subscribers of a channel
- Requires `channel` field
- Type: `CHANNEL`

## Getting Started

1. **Choose your SDK**: Select Go or Python based on your project requirements
2. **Install dependencies**: Follow installation instructions in the respective SDK documentation
3. **Generate protobuf code**: (if needed) Generate language-specific code from the `.proto` file
4. **Connect to hub**: Create a client instance and connect to the server
5. **Send/receive messages**: Use SDK methods to interact with the hub

## Server Configuration

Default server configuration:
- **Host**: localhost
- **Port**: 50051
- **Protocol**: gRPC (insecure for development)

## Common Patterns

### 1. Simple Client

```
1. Connect to server
2. Send a message
3. Close connection
```

### 2. Long-lived Client

```
1. Connect to server
2. Start message receiver (in separate thread/goroutine)
3. Send messages as needed
4. Handle disconnections with reconnection logic
```

### 3. Pub/Sub Pattern

```
1. Connect to server
2. Subscribe to specific channels
3. Publish messages to channels
4. Receive messages from subscribed channels
```

## Best Practices

1. **Error Handling**: Always handle connection and transmission errors
2. **Resource Cleanup**: Use defer (Go) or context managers (Python) to ensure proper cleanup
3. **Concurrency**: Run message receivers in separate threads/goroutines
4. **Reconnection**: Implement automatic reconnection for production applications
5. **Logging**: Add appropriate logging for debugging and monitoring

## Examples by Use Case

### Chat Application
- Use **broadcast** for public messages
- Use **direct** for private messages
- Use **channels** for group conversations

### Notification System
- Use **broadcast** for system-wide alerts
- Use **direct** for user-specific notifications
- Use **channels** for topic-based notifications

### Real-time Updates
- Use **channels** for different data streams
- Subscribe to relevant channels based on user preferences

## Support

For more information:
- Check the SDK-specific documentation
- Review example code in each SDK directory
- See the main project README at `/README.md`

## Contributing

To add support for additional languages:
1. Create a new directory under `/docs/sdk/`
2. Generate protobuf code for the target language
3. Create a client wrapper library
4. Write comprehensive documentation with examples
5. Add tests and usage examples
