package client

import (
	pb "borrowing-management-services/proto/book"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BookGrpcClientInterface interface {
	GetBookStock(ctx context.Context, bookID string) (int32, int32, error)
	Close() error
}

type BookGrpcClient struct {
	conn   *grpc.ClientConn
	client pb.BookStockServiceClient
}

func NewBookGrpcClient(targetHost string) (BookGrpcClientInterface, error) {

	conn, err := grpc.NewClient(targetHost, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	return &BookGrpcClient{
		conn:   conn,
		client: pb.NewBookStockServiceClient(conn),
	}, nil

}

func (c *BookGrpcClient) GetBookStock(ctx context.Context, bookID string) (int32, int32, error) {

	res, err := c.client.GetBookStock(ctx, &pb.GetBookStockRequest{BookId: bookID})

	if err != nil {
		return 0, 0, err
	}

	return res.GetAvailableStock(), res.GetTotalStock(), nil

}

func (c *BookGrpcClient) Close() error {
	return c.conn.Close()
}
