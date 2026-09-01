package grpc

import (
	"log"
	"net"
	"user-management-services/internal/usecase"
	pb "user-management-services/proto/user"

	"google.golang.org/grpc"
)

type GrpcServer struct {
	server   *grpc.Server
	listener net.Listener
}

func NewGrpcServer(port string, usecase usecase.UserProfileUsecaseInterface) (*GrpcServer, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	handler := NewUserStatusGRPCHandler(usecase)

	pb.RegisterUserStatusServiceServer(grpcServer, handler)

	return &GrpcServer{
		server:   grpcServer,
		listener: lis,
	}, nil
}

func (s *GrpcServer) Start() {
	go func() {
		log.Printf("🚀 gRPC Server running on port %s", s.listener.Addr().String())
		if err := s.server.Serve(s.listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()
}

func (s *GrpcServer) Stop() {
	s.server.GracefulStop()
}
