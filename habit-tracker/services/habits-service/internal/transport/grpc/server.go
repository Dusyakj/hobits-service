package grpc

import (
	"fmt"
	"net"

	pb "habits-service/proto/habits/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server represents a gRPC server
type Server struct {
	grpcServer *grpc.Server
	handler    *HabitServiceHandler
	port       int
}

// NewServer creates a new gRPC server
func NewServer(handler *HabitServiceHandler, port int) *Server {
	grpcServer := grpc.NewServer(
	// TODO: Add interceptors for logging, metrics, recovery
	)

	pb.RegisterHabitServiceServer(grpcServer, handler)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

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
