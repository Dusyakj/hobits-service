package main

import (
	"habits-service/internal/app"
	"log"
)

func main() {
	// Create and initialize the application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
