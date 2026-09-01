package client

import (
	pb "borrowing-management-services/proto/user"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserGrpcClientInterface interface {
	GetUserStatus(ctx context.Context, userID string) (string, error)
	Close() error
}

type UserGrpcClient struct {
	conn   *grpc.ClientConn
	client pb.UserStatusServiceClient
}

func NewUserGrpcClient(targetHost string) (UserGrpcClientInterface, error) {

	conn, err := grpc.NewClient(targetHost, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewUserStatusServiceClient(conn)

	return &UserGrpcClient{
		conn:   conn,
		client: client,
	}, nil

}

func (c *UserGrpcClient) GetUserStatus(ctx context.Context, userID string) (string, error) {
	req := &pb.GetUserStatusRequest{
		UserId: userID,
	}

	res, err := c.client.GetUserStatus(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetStatus(), nil
}

func (c *UserGrpcClient) Close() error {
	return c.conn.Close()
}
