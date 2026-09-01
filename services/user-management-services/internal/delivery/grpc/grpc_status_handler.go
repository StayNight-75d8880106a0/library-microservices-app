package grpc

import (
	"context"
	"user-management-services/internal/usecase"
	pb "user-management-services/proto/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserStatusGRPCHandler struct {
	pb.UnimplementedUserStatusServiceServer
	usecase usecase.UserProfileUsecaseInterface
}

func NewUserStatusGRPCHandler(usecase usecase.UserProfileUsecaseInterface) *UserStatusGRPCHandler {
	return &UserStatusGRPCHandler{
		usecase: usecase,
	}
}

func (h *UserStatusGRPCHandler) GetUserStatus(ctx context.Context, request *pb.GetUserStatusRequest) (*pb.GetUserStatusResponse, error) {

	if request.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID cannot be empty")
	}

	userStatus, err := h.usecase.GetUserStatus(ctx, request.GetUserId())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.GetUserStatusResponse{
		UserId: request.GetUserId(),
		Status: userStatus,
	}

	return result, nil

}
