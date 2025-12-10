package hub

import (
	"sync"
	"time"
)

// RequestInfo stores information about a pending request
type RequestInfo struct {
	RequestID   string
	RequesterID string // Original client who made the request
	WorkerID    string // Worker processing the request
	Capability  string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

// RequestTracker tracks active requests and routes responses back
type RequestTracker struct {
	mu       sync.RWMutex
	requests map[string]*RequestInfo // request_id -> RequestInfo
}

// NewRequestTracker creates a new request tracker
func NewRequestTracker() *RequestTracker {
	tracker := &RequestTracker{
		requests: make(map[string]*RequestInfo),
	}
	
	// Start cleanup goroutine
	go tracker.cleanupExpired()
	
	return tracker
}

// Track registers a new request
func (rt *RequestTracker) Track(requestID, requesterID, workerID, capability string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	
	rt.requests[requestID] = &RequestInfo{
		RequestID:   requestID,
		RequesterID: requesterID,
		WorkerID:    workerID,
		Capability:  capability,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(5 * time.Minute), // 5 minute timeout
	}
}

// GetRequester retrieves the original requester for a request_id
func (rt *RequestTracker) GetRequester(requestID string) (requesterID string, found bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	
	if info, exists := rt.requests[requestID]; exists {
		return info.RequesterID, true
	}
	return "", false
}

// Complete removes a request from tracking
func (rt *RequestTracker) Complete(requestID string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	
	delete(rt.requests, requestID)
}

// cleanupExpired removes expired requests periodically
func (rt *RequestTracker) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rt.mu.Lock()
		now := time.Now()
		for requestID, info := range rt.requests {
			if now.After(info.ExpiresAt) {
				delete(rt.requests, requestID)
			}
		}
		rt.mu.Unlock()
	}
}

// GetStats returns current tracking statistics
func (rt *RequestTracker) GetStats() map[string]interface{} {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	
	return map[string]interface{}{
		"active_requests": len(rt.requests),
	}
}
