package cron

import (
	"context"
	"fmt"
	"habits-service/internal/domain/service"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// DeadlineChecker periodically checks for missed deadlines and resets streaks
type DeadlineChecker struct {
	habitService service.HabitService
	cron         *cron.Cron
	interval     time.Duration
}

// NewDeadlineChecker creates a new deadline checker
func NewDeadlineChecker(habitService service.HabitService, checkInterval time.Duration) *DeadlineChecker {
	return &DeadlineChecker{
		habitService: habitService,
		cron:         cron.New(),
		interval:     checkInterval,
	}
}

// Start starts the deadline checker
func (d *DeadlineChecker) Start() error {
	cronExpr := fmt.Sprintf("@every %s", d.interval.String())

	log.Printf("Starting deadline checker with interval: %s", d.interval)

	_, err := d.cron.AddFunc(cronExpr, func() {
		d.checkDeadlines()
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	d.cron.Start()
	log.Println("Deadline checker started successfully")

	return nil
}

// Stop stops the deadline checker
func (d *DeadlineChecker) Stop() {
	log.Println("Stopping deadline checker...")
	ctx := d.cron.Stop()
	<-ctx.Done()
	log.Println("Deadline checker stopped")
}

// checkDeadlines runs the deadline check logic
func (d *DeadlineChecker) checkDeadlines() {
	log.Println("Running deadline check...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Reset confirmation flags for habits entering new period
	err := d.habitService.ResetConfirmationFlags(ctx)
	if err != nil {
		log.Printf("Error resetting confirmation flags: %v", err)
	}

	// Process missed deadlines and reset streaks
	err = d.habitService.ProcessMissedDeadlines(ctx)
	if err != nil {
		log.Printf("Error processing missed deadlines: %v", err)
		return
	}

	log.Println("Deadline check completed successfully")
}
