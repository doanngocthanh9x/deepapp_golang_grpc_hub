package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "deepapp_golang_grpc_hub/internal/proto"
)

// HubClient represents the gRPC hub client
type HubClient struct {
	conn      *grpc.ClientConn
	client    pb.HubServiceClient
	stream    pb.HubService_ConnectClient
	ClientID  string // Exported for access
	responses chan *pb.Message
}

// NewHubClient creates a new hub client
func NewHubClient(serverAddr string) (*HubClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	client := pb.NewHubServiceClient(conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	hc := &HubClient{
		conn:      conn,
		client:    client,
		stream:    stream,
		ClientID:  fmt.Sprintf("web-api-%d", time.Now().UnixNano()),
		responses: make(chan *pb.Message, 100),
	}

	// Start receiving messages
	go hc.receiveMessages()

	return hc, nil
}

func (hc *HubClient) receiveMessages() {
	for {
		msg, err := hc.stream.Recv()
		if err != nil {
			log.Printf("Receive error: %v", err)
			return
		}
		hc.responses <- msg
	}
}

// SendRequest sends a request to the hub
func (hc *HubClient) SendRequest(targetWorker, capability, data string) (*pb.Message, error) {
	msg := pb.Message{
		Id:        fmt.Sprintf("req-%d", time.Now().UnixNano()),
		From:      hc.ClientID,
		To:        targetWorker,
		Content:   data,
		Channel:   capability, // Use capability as channel for worker routing
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      pb.MessageType_REQUEST,
		Action:    "request",
		Metadata: map[string]string{
			"capability": capability,
		},
	}

	log.Printf("📤 Sending request: Type=%v (%d), Action='%s', Capability='%s', To='%s'",
		msg.Type, msg.Type, msg.Action, capability, targetWorker)

	if err := hc.stream.Send(&msg); err != nil {
		return nil, err
	}

	// Wait for response with timeout
	select {
	case response := <-hc.responses:
		return response, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Close closes the hub client connection
func (hc *HubClient) Close() error {
	if hc.conn != nil {
		return hc.conn.Close()
	}
	return nil
}