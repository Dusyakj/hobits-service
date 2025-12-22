package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"habits-service/internal/config"
	cronpkg "habits-service/internal/infrastructure/cron"
	infradb "habits-service/internal/infrastructure/db"
	"habits-service/internal/infrastructure/postgres"
	"habits-service/internal/service"
	"habits-service/internal/transport/grpc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// App represents the application
type App struct {
	config          *config.Config
	grpcServer      *grpc.Server
	deadlineChecker *cronpkg.DeadlineChecker
	dbPool          *pgxpool.Pool
}

// New creates a new application
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Configuration loaded successfully")

	// Initialize PostgreSQL connection pool
	ctx := context.Background()
	dbPool, err := infradb.NewPostgresPool(ctx, &cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	fmt.Println("Connected to PostgreSQL")

	// Initialize repositories
	habitRepo := postgres.NewHabitRepository(dbPool)
	confirmationRepo := postgres.NewHabitConfirmationRepository(dbPool)

	// Initialize services
	habitService := service.NewHabitService(habitRepo, confirmationRepo)
	fmt.Println("Services initialized")

	// Initialize deadline checker (if enabled)
	var deadlineChecker *cronpkg.DeadlineChecker
	if cfg.Scheduler.Enabled {
		deadlineChecker = cronpkg.NewDeadlineChecker(habitService, cfg.Scheduler.CheckInterval)
		fmt.Println("Deadline checker initialized")
	} else {
		fmt.Println("Deadline checker is disabled in configuration")
	}

	// Initialize gRPC handler
	grpcHandler := grpc.NewHabitServiceHandler(habitService)

	// Initialize gRPC server
	grpcServer := grpc.NewServer(grpcHandler, cfg.GRPC.Port)

	return &App{
		config:          cfg,
		grpcServer:      grpcServer,
		deadlineChecker: deadlineChecker,
		dbPool:          dbPool,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	// Setup signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start deadline checker if enabled
	if a.deadlineChecker != nil {
		if err := a.deadlineChecker.Start(); err != nil {
			return fmt.Errorf("failed to start deadline checker: %w", err)
		}
	}

	// Start gRPC server in a goroutine
	go func() {
		if err := a.grpcServer.Start(); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
			quit <- syscall.SIGTERM
		}
	}()

	fmt.Printf("%s service started on port %d\n", a.config.Service.Name, a.config.GRPC.Port)
	fmt.Println("Press Ctrl+C to shutdown...")

	// Wait for interrupt signal
	<-quit
	fmt.Println("\nShutting down server...")

	// Graceful shutdown
	a.grpcServer.Stop()

	// Stop deadline checker
	if a.deadlineChecker != nil {
		a.deadlineChecker.Stop()
	}

	// Close database pool
	a.dbPool.Close()

	fmt.Println("Server shutdown complete")
	return nil
}
