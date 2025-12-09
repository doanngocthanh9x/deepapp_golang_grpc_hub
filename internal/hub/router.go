package hub

import (
	"deepapp_golang_grpc_hub/internal/proto"
)

type Router struct {
	connMgr *ConnectionManager
	subMgr  *SubscriberManager
}

func NewRouter(connMgr *ConnectionManager, subMgr *SubscriberManager) *Router {
	return &Router{
		connMgr: connMgr,
		subMgr:  subMgr,
	}
}

func (r *Router) Route(msg *proto.Message) {
	switch msg.Type {
	case proto.MessageType_DIRECT:
		r.routeDirect(msg)
	case proto.MessageType_BROADCAST:
		r.routeBroadcast(msg)
	case proto.MessageType_CHANNEL:
		r.routeChannel(msg)
	case proto.MessageType_REQUEST, proto.MessageType_RESPONSE:
		// Route requests and responses as direct messages
		r.routeDirect(msg)
	}
}

func (r *Router) routeDirect(msg *proto.Message) {
	if stream, exists := r.connMgr.Get(msg.To); exists {
		stream.Send(msg)
	}
}

func (r *Router) routeBroadcast(msg *proto.Message) {
	r.connMgr.Broadcast(msg)
}

func (r *Router) routeChannel(msg *proto.Message) {
	r.subMgr.Publish(msg.Channel, msg)
}