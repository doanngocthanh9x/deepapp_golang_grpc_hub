package hub

import (
	"sync"

	"deepapp_golang_grpc_hub/internal/proto"
)

type SubscriberManager struct {
	mu          sync.RWMutex
	subscribers map[string][]proto.HubService_ConnectServer
}

func NewSubscriberManager() *SubscriberManager {
	return &SubscriberManager{
		subscribers: make(map[string][]proto.HubService_ConnectServer),
	}
}

func (sm *SubscriberManager) Subscribe(channel string, stream proto.HubService_ConnectServer) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.subscribers[channel] = append(sm.subscribers[channel], stream)
}

func (sm *SubscriberManager) Unsubscribe(channel string, stream proto.HubService_ConnectServer) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	subs := sm.subscribers[channel]
	for i, s := range subs {
		if s == stream {
			sm.subscribers[channel] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

func (sm *SubscriberManager) Publish(channel string, msg *proto.Message) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, stream := range sm.subscribers[channel] {
		stream.Send(msg)
	}
}