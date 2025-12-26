package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"user-service/internal/config"
	infradb "user-service/internal/infrastructure/db"
	"user-service/internal/infrastructure/kafka"
	"user-service/internal/infrastructure/postgres"
	infraredis "user-service/internal/infrastructure/redis"
	"user-service/internal/service"
	"user-service/internal/transport/grpc"
	"user-service/pkg/jwt"
)

// App represents the application
type App struct {
	config        *config.Config
	grpcServer    *grpc.Server
	kafkaProducer *kafka.Producer
}

// New creates a new application
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Configuration loaded successfully")

	ctx := context.Background()
	pgPool, err := infradb.NewPostgresPool(ctx, &cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	fmt.Println("Connected to PostgreSQL")

	redisClient, err := infraredis.NewSessionRedisClient(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	fmt.Println("Connected to Redis")

	userRepo := postgres.NewUserRepository(pgPool)
	sessionRepo := postgres.NewSessionRepository(pgPool)

	sessionStorage := infraredis.NewSessionStorage(redisClient, cfg.Redis.SessionTTL)

	verificationTokenStorage := infraredis.NewVerificationTokenStorage(redisClient)

	passwordResetTokenStorage := infraredis.NewPasswordResetTokenStorage(redisClient)

	kafkaProducer := kafka.NewProducer(&cfg.Kafka)
	fmt.Println("Kafka producer initialized")

	tokenManager := jwt.NewTokenManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.Issuer,
	)

	// Initialize services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(
		userService,
		sessionRepo,
		sessionStorage,
		verificationTokenStorage,
		passwordResetTokenStorage,
		tokenManager,
		kafkaProducer,
	)

	grpcHandler := grpc.NewUserServiceHandler(userService, authService)

	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPC.Port)

	return &App{
		config:        cfg,
		grpcServer:    grpcServer,
		kafkaProducer: kafkaProducer,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := a.grpcServer.Start(); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
			quit <- syscall.SIGTERM
		}
	}()

	fmt.Printf("User service started on port %d\n", a.config.GRPC.Port)
	fmt.Println("Press Ctrl+C to shutdown...")

	<-quit
	fmt.Println("\nShutting down server...")

	a.grpcServer.Stop()

	if err := a.kafkaProducer.Close(); err != nil {
		fmt.Printf("Error closing Kafka producer: %v\n", err)
	}

	fmt.Println("Server shutdown complete")
	return nil
}
