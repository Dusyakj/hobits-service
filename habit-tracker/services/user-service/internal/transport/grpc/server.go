package grpc

import (
	"fmt"
	"net"

	pb "user-service/proto/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents a gRPC server
type Server struct {
	grpcServer *grpc.Server
	handler    *UserServiceHandler
	port       int
}

// NewServer creates a new gRPC server
func NewServer(handler *UserServiceHandler, port int) *Server {
	grpcServer := grpc.NewServer(
	// TODO: Add interceptors for logging, metrics, recovery
	)

	pb.RegisterUserServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		handler:    handler,
		port:       port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	fmt.Printf("gRPC server listening on :%d\n", s.port)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	fmt.Println("Gracefully stopping gRPC server...")
	s.grpcServer.GracefulStop()
	fmt.Println("gRPC server stopped")
}
