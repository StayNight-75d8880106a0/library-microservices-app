package grpc

import (
	"context"
	"errors"
	"user-management-services/internal/helper"
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
		var appErr *helper.AppError
		if errors.As(err, &appErr) {
			return nil, status.Error(httpToGRPCCode(appErr.Code), appErr.ErrorMessage)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.GetUserStatusResponse{
		UserId: request.GetUserId(),
		Status: userStatus,
	}

	return result, nil

}

func httpToGRPCCode(httpCode int) codes.Code {
	switch httpCode {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 408:
		return codes.DeadlineExceeded
	case 409:
		return codes.AlreadyExists
	case 429:
		return codes.ResourceExhausted
	case 501:
		return codes.Unimplemented
	case 503:
		return codes.Unavailable
	case 504:
		return codes.DeadlineExceeded
	default:
		if httpCode >= 500 {
			return codes.Internal
		}
		return codes.Unknown
	}
}
