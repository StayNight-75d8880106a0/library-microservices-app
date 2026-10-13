package grpc

import (
	"book-management-services/internal/repository"
	pb "book-management-services/proto/book"
	"log"
	"net"

	"google.golang.org/grpc"
)

type GrpcServer struct {
	server   *grpc.Server
	listener net.Listener
}

func NewGrpcServer(port string, bookRepository repository.BookRepositoryInterface) (*GrpcServer, error) {

	lis, err := net.Listen("tcp", ":"+port)

	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	pb.RegisterBookStockServiceServer(grpcServer, NewBookStockGRPCHandler(bookRepository))

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
