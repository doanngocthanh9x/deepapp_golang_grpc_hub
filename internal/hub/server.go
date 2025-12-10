package hub

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"deepapp_golang_grpc_hub/internal/config"
	"deepapp_golang_grpc_hub/internal/proto"
)

type Server struct {
	config     *config.Config
	server     *grpc.Server
	connMgr    *ConnectionManager
	router     *Router
	subMgr     *SubscriberManager
	dispatcher *Dispatcher
	handler    *Handler
	registry   *ServiceRegistry // Service registry with DB persistence
}

func NewServer(cfg *config.Config) *Server {
	fmt.Println("Creating ConnectionManager...")
	connMgr := NewConnectionManager()
	fmt.Println("Creating SubscriberManager...")
	subMgr := NewSubscriberManager()
	fmt.Println("Creating ServiceRegistry...")
	registry := NewServiceRegistry()
	fmt.Println("Creating Router...")
	router := NewRouter(connMgr, subMgr)
	fmt.Println("Creating Dispatcher...")
	dispatcher := NewDispatcher(router)
	fmt.Println("Creating Handler...")
	handler := NewHandler(nil) // TODO: add repo

	fmt.Println("Creating gRPC server...")
	s := &Server{
		config:     cfg,
		server:     grpc.NewServer(),
		connMgr:    connMgr,
		router:     router,
		subMgr:     subMgr,
		dispatcher: dispatcher,
		handler:    handler,
		registry:   registry,
	}

	fmt.Println("Registering HubService...")
	proto.RegisterHubServiceServer(s.server, s)
	fmt.Println("Registering reflection...")
	reflection.Register(s.server)

	fmt.Println("Server fully initialized")
	return s
}

func NewServerWithRegistry(cfg *config.Config, registry *ServiceRegistry) *Server {
	fmt.Println("Creating ConnectionManager...")
	connMgr := NewConnectionManager()
	fmt.Println("Creating SubscriberManager...")
	subMgr := NewSubscriberManager()
	fmt.Println("Creating Router...")
	router := NewRouter(connMgr, subMgr)
	fmt.Println("Creating Dispatcher...")
	dispatcher := NewDispatcher(router)
	fmt.Println("Creating Handler...")
	handler := NewHandler(nil)

	fmt.Println("Creating gRPC server...")
	s := &Server{
		config:     cfg,
		server:     grpc.NewServer(),
		connMgr:    connMgr,
		router:     router,
		subMgr:     subMgr,
		dispatcher: dispatcher,
		handler:    handler,
		registry:   registry,
	}

	fmt.Println("Registering HubService...")
	proto.RegisterHubServiceServer(s.server, s)
	fmt.Println("Registering reflection...")
	reflection.Register(s.server)

	fmt.Println("Server fully initialized with custom registry")
	return s
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.config.Port)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Server is now listening on port %s\n", s.config.Port)
	fmt.Println("Server is ready to accept connections...")
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) Connect(stream proto.HubService_ConnectServer) error {
	// Wait for first message to get client ID
	firstMsg, err := stream.Recv()
	if err != nil {
		return err
	}

	clientID := firstMsg.From
	if clientID == "" {
		clientID = "client-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	fmt.Printf("✓ Client connected: %s\n", clientID)
	s.connMgr.Add(clientID, stream)
	defer func() {
		s.connMgr.Remove(clientID)
		s.registry.UnregisterWorker(clientID)
		fmt.Printf("✗ Client disconnected: %s\n", clientID)
	}()

	// Process first message (could be registration)
	s.handleMessage(firstMsg)

	// Continue receiving messages
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		fmt.Printf("→ Message from %s to %s (type: %v)\n", msg.From, msg.To, msg.Type)
		s.handleMessage(msg)
	}
}

func (s *Server) handleMessage(msg *proto.Message) {
	// Handle registration messages
	if msg.Type == proto.MessageType_REGISTER {
		s.handleRegistration(msg)
		return
	}

	// Handle worker-to-worker calls
	if msg.Type == proto.MessageType_WORKER_CALL {
		s.handleWorkerCall(msg)
		return
	}

	// Handle responses (including worker-to-worker responses)
	if msg.Type == proto.MessageType_RESPONSE {
		s.handleResponse(msg)
		return
	}

	// Handle capability discovery requests
	if msg.Channel == "capability_discovery" || (msg.Type == proto.MessageType_REQUEST && msg.Content != "") {
		var reqData map[string]interface{}
		if err := json.Unmarshal([]byte(msg.Content), &reqData); err == nil {
			if action, ok := reqData["action"].(string); ok {
				if action == "discover" || action == "list_capabilities" {
					s.handleCapabilityDiscovery(msg)
					return
				}
			}
		}
		// Also check channel
		if msg.Channel == "capability_discovery" {
			s.handleCapabilityDiscovery(msg)
			return
		}
	}

	// Handle regular request routing
	if msg.Type == proto.MessageType_REQUEST {
		s.handleServiceRequest(msg)
		return
	}

	// Default: dispatch to router
	s.dispatcher.Dispatch(msg)
}