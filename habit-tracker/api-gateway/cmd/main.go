package main

import (
	"log"

	_ "api-gateway/docs"
	"api-gateway/internal/app"
)

// @title Habit Tracker API
// @version 1.0
// @description API Gateway for Habit Tracker microservices
// @description Provides REST API for user management, habits tracking, and authentication

// @contact.name
// @contact.email

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
