package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/config"
	"notification-service/internal/infrastructure/db"
	"notification-service/internal/infrastructure/kafka"
	"notification-service/internal/infrastructure/postgres"
	"notification-service/internal/infrastructure/redis"
	"notification-service/internal/infrastructure/smtp"
	"notification-service/internal/service"
)

type App struct {
	cfg *config.Config
}

// New creates a new application instance
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &App{
		cfg: cfg,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	ctx := context.Background()

	// Initialize PostgreSQL
	log.Println("Connecting to PostgreSQL...")
	pool, err := db.NewPostgresPool(ctx, &a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close(pool)
	log.Println("Connected to PostgreSQL")

	// Initialize Redis
	log.Println("Connecting to Redis...")
	redisClient, err := redis.NewRedisClient(&a.cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redis.Close(redisClient)
	log.Println("Connected to Redis")

	// Initialize SMTP client
	log.Println("Initializing SMTP client...")
	smtpClient, err := smtp.NewClient(&a.cfg.SMTP, &a.cfg.Email)
	if err != nil {
		return fmt.Errorf("failed to initialize SMTP client: %w", err)
	}
	log.Println("SMTP client initialized")

	// Initialize repositories
	notificationRepo := postgres.NewNotificationRepository(pool)

	// Initialize services
	emailService := service.NewEmailService(smtpClient)
	notificationService := service.NewNotificationService(notificationRepo, emailService)

	// Initialize Kafka consumer
	log.Println("Initializing Kafka consumer...")
	consumer := kafka.NewConsumer(&a.cfg.Kafka, notificationService, emailService)
	log.Println("Kafka consumer initialized")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start Kafka consumer in goroutine
	consumerErrChan := make(chan error, 1)
	go func() {
		if err := consumer.Start(ctx); err != nil {
			consumerErrChan <- err
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Notification service started successfully")

	select {
	case err := <-consumerErrChan:
		log.Printf("Kafka consumer error: %v", err)
		return err
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()

		// Graceful shutdown
		log.Println("Shutting down gracefully...")
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing Kafka consumer: %v", err)
		}
	}

	log.Println("Application stopped")
	return nil
}
