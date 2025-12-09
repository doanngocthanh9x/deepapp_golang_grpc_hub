package hub

import (
	"sync"

	"deepapp_golang_grpc_hub/internal/proto"
)

type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]proto.HubService_ConnectServer
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]proto.HubService_ConnectServer),
	}
}

func (cm *ConnectionManager) Add(clientID string, stream proto.HubService_ConnectServer) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.connections[clientID] = stream
}

func (cm *ConnectionManager) Remove(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, clientID)
}

func (cm *ConnectionManager) Get(clientID string) (proto.HubService_ConnectServer, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	stream, exists := cm.connections[clientID]
	return stream, exists
}

func (cm *ConnectionManager) Broadcast(msg *proto.Message) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for _, stream := range cm.connections {
		stream.Send(msg)
	}
}