package grpc

import (
	"book-management-services/internal/repository"
	pb "book-management-services/proto/book"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type BookStockGRPCHandler struct {
	pb.UnimplementedBookStockServiceServer
	repository repository.BookRepositoryInterface
}

func NewBookStockGRPCHandler(repository repository.BookRepositoryInterface) *BookStockGRPCHandler {
	return &BookStockGRPCHandler{
		repository: repository,
	}
}

func (h *BookStockGRPCHandler) GetBookStock(ctx context.Context, request *pb.GetBookStockRequest) (*pb.GetBookStockResponse, error) {

	if request.GetBookId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Book ID cannot be empty")
	}

	book, errGet := h.repository.GetById(ctx, request.GetBookId())

	if errGet != nil {
		if errors.Is(errGet, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "Book not found")
		}
		return nil, status.Error(codes.Internal, "Failed to get book stock")
	}

	result := &pb.GetBookStockResponse{
		BookId:         book.ID,
		AvailableStock: int32(book.AvailableStock),
		TotalStock:     int32(book.TotalStock),
	}

	return result, nil

}
