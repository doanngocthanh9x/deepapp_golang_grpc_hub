package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"deepapp_golang_grpc_hub/internal/proto"
	"deepapp_golang_grpc_hub/internal/utils"
)

func main() {
	// Connect to server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewHubServiceClient(conn)

	// Start streaming
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	// Generate client ID
	clientID := utils.GenerateID()
	fmt.Printf("Connected as client: %s\n", clientID)

	// Goroutine to receive messages
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("Stream error: %v", err)
				return
			}
			fmt.Printf("Received: %s\n", msg.Content)
		}
	}()

	// Send messages from stdin
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter messages (format: type:to:content or 'broadcast:content' or 'channel:chan:content'):")

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 2 {
			fmt.Println("Invalid format")
			continue
		}

		msg := proto.Message{
			Id:        utils.GenerateID(),
			From:      clientID,
			Timestamp: utils.FormatTime(utils.Now()),
		}

		switch parts[0] {
		case "direct":
			if len(parts) != 3 {
				fmt.Println("Direct message format: direct:to:content")
				continue
			}
			msg.Type = proto.MessageType_DIRECT
			msg.To = parts[1]
			msg.Content = parts[2]
		case "broadcast":
			msg.Type = proto.MessageType_BROADCAST
			msg.Content = parts[1]
		case "channel":
			if len(parts) != 3 {
				fmt.Println("Channel message format: channel:chan:content")
				continue
			}
			msg.Type = proto.MessageType_CHANNEL
			msg.Channel = parts[1]
			msg.Content = parts[2]
		default:
			fmt.Println("Unknown type. Use: direct, broadcast, or channel")
			continue
		}

		if err := stream.Send(&msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}