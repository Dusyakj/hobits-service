package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"
	habitspb "api-gateway/proto/habits/v1"
	userpb "api-gateway/proto/user/v1"
)

// App represents the application
type App struct {
	cfg        *config.Config
	httpServer *http.Server
	grpcConns  []*grpc.ClientConn
}

// New creates a new application
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	app := &App{
		cfg:       cfg,
		grpcConns: make([]*grpc.ClientConn, 0),
	}

	if err := app.initGRPCClients(); err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC clients: %w", err)
	}

	if err := app.initHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	return app, nil
}

// initGRPCClients initializes connections to all gRPC services
func (a *App) initGRPCClients() error {
	userConn, err := grpc.NewClient(
		a.cfg.GRPC.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to user-service: %w", err)
	}
	a.grpcConns = append(a.grpcConns, userConn)

	habitsConn, err := grpc.NewClient(
		a.cfg.GRPC.HabitsServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to habits-service: %w", err)
	}
	a.grpcConns = append(a.grpcConns, habitsConn)

	// TODO: Connect to bad-habits-service when ready
	// badHabitsConn, err := grpc.NewClient(...)

	log.Printf("Connected to gRPC services")
	return nil
}

// initHTTPServer initializes the HTTP server with all handlers and middleware
func (a *App) initHTTPServer() error {
	userClient := userpb.NewUserServiceClient(a.grpcConns[0])
	habitsClient := habitspb.NewHabitServiceClient(a.grpcConns[1])

	authMiddleware := middleware.NewAuthMiddleware(userClient)

	userHandler := handler.NewUserHandler(userClient)
	habitHandler := handler.NewHabitHandler(habitsClient)

	router := handler.NewRouter(userHandler, habitHandler, authMiddleware)
	httpHandler := router.Setup()

	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.cfg.HTTP.Port),
		Handler:      httpHandler,
		ReadTimeout:  time.Duration(a.cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.cfg.HTTP.WriteTimeout) * time.Second,
	}

	log.Printf("HTTP server configured on port %d", a.cfg.HTTP.Port)
	return nil
}

// Run starts the application
func (a *App) Run() error {
	log.Println("Starting rate limit cleanup routine")
	middleware.CleanupVisitors()

	go func() {
		log.Printf("Starting HTTP server on %s", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	for _, conn := range a.grpcConns {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}

	log.Println("Server stopped")
	return nil
}
