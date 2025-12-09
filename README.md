# DeepApp Golang gRPC Hub

A gRPC-based hub service for real-time messaging with support for direct messages, broadcasts, and channel-based communication.

## Features

- gRPC streaming for real-time communication
- Direct messaging between clients
- Broadcast messages to all connected clients
- Channel-based messaging with subscriptions
- SQLite persistence for messages and client states
- Configurable via environment variables

## Project Structure

```
deepapp_golang_grpc_hub/
├── cmd/
│   └── hub/
│       └── main.go                  # Entry point chạy server
├── internal/
│   ├── config/
│   │   └── config.go                # Load env, config app
│   ├── hub/
│   │   ├── server.go                # gRPC Stream implementation
│   │   ├── connection.go            # Connection manager (map client->stream)
│   │   ├── router.go                # Route direct, broadcast, channel publish
│   │   ├── subscriber.go            # Channel subscriber manager
│   │   ├── dispatcher.go            # Queue dispatch (optional)
│   │   └── handler.go               # Handle Request types: json/file/control
│   ├── proto/                       # Generated code từ .proto
│   │   └── hub.pb.go
│   ├── repository/
│   │   ├── messages_repo.go         # Lưu & đọc message từ SQLite
│   │   └── clients_repo.go          # Lưu trạng thái client (optional persist)
│   ├── models/
│   │   ├── message.go               # Struct để bind DB
│   │   └── client.go
│   ├── db/
│   │   ├── sqlite.go                # Init + kết nối DB
│   │   └── migrations/
│   │       ├── 001_create_messages.sql
│   │       └── 002_create_clients.sql
│   └── utils/
│       ├── json.go                  # Convert Struct <-> JSON
│       ├── id.go                    # message_id generator (uuid)
│       └── time.go
├── pkg/
│   └── logger/                      # Logger singleton (zap/logrus)
│       └── logger.go
├── proto/
│   └── hub.proto                    # File proto nguồn
├── Makefile                         # generate proto, run server
├── go.mod
└── README.md
```

## Setup

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Generate protobuf code:
   ```bash
   make proto
   ```

3. Run the server:
   ```bash
   make run
   ```

## Configuration

Configure the application using environment variables:

- `PORT`: Server port (default: 50051)
- `LOG_LEVEL`: Logging level (default: info)
- `DB_PATH`: SQLite database path (default: hub.db)

## Usage

Connect to the gRPC service using the defined proto interface. The service supports streaming messages with different types: DIRECT, BROADCAST, and CHANNEL.

## Client Usage

### Connecting to the Hub

Clients connect to the hub using gRPC streaming. Each client gets a unique ID upon connection.

### Running the Example Client

An example client is provided in `cmd/client/main.go`. To run it:

```bash
make run-client
```

or

```bash
go run cmd/client/main.go
```

### Message Types

The hub supports three types of messages:

1. **Direct Messages**: Send to a specific client

   ```text
   direct:<target_client_id>:<message_content>
   ```

2. **Broadcast Messages**: Send to all connected clients

   ```text
   broadcast:<message_content>
   ```

3. **Channel Messages**: Send to all subscribers of a channel

   ```text
   channel:<channel_name>:<message_content>
   ```

### Example Usage

After starting the server and running the client, you can send messages like:

- `broadcast:Hello everyone!`
- `direct:client-123:Private message`
- `channel:news:Breaking news!`

The client will receive messages in real-time through the stream.
