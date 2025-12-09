package hub

import (
	"context"

	"deepapp_golang_grpc_hub/internal/proto"
	"deepapp_golang_grpc_hub/internal/repository"
	"deepapp_golang_grpc_hub/pkg/logger"
)

type Handler struct {
	repo *repository.MessagesRepo
}

func NewHandler(repo *repository.MessagesRepo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Handle(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	switch req.Type {
	case proto.RequestType_JSON:
		return h.handleJSON(req)
	case proto.RequestType_FILE:
		return h.handleFile(req)
	case proto.RequestType_CONTROL:
		return h.handleControl(req)
	default:
		logger.Warn("Unknown request type")
		return &proto.Response{Status: proto.Status_ERROR}, nil
	}
}

func (h *Handler) handleJSON(req *proto.Request) (*proto.Response, error) {
	// Process JSON request
	logger.Info("Handling JSON request")
	return &proto.Response{Status: proto.Status_OK}, nil
}

func (h *Handler) handleFile(req *proto.Request) (*proto.Response, error) {
	// Process file request
	logger.Info("Handling file request")
	return &proto.Response{Status: proto.Status_OK}, nil
}

func (h *Handler) handleControl(req *proto.Request) (*proto.Response, error) {
	// Process control request
	logger.Info("Handling control request")
	return &proto.Response{Status: proto.Status_OK}, nil
}