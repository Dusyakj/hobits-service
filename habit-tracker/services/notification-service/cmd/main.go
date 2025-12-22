package main

import (
	"log"

	"notification-service/internal/app"
)

func main() {
	// Create application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Run application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
