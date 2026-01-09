package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	"github.com/weni-ai/flows-code-actions/internal/db/postgre"
	log "github.com/sirupsen/logrus"
)

// Example demonstrates how to use PostgreSQL coderun repository
func Example() error {
	cfg := &config.Config{
		DB: config.DBConfig{
			URI:         "postgres://localhost:5432/codeactions?sslmode=disable",
			Name:        "codeactions",
			Timeout:     30,
			MaxPoolSize: 25,
			MinPoolSize: 5,
		},
	}

	// Connect to PostgreSQL
	db, err := postgre.GetPostgreDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	// Create repository
	coderunRepo := NewCodeRunRepository(db)

	ctx := context.Background()

	// Create a new coderun
	newCodeRun := coderun.NewCodeRun("507f1f77bcf86cd799439011", coderun.StatusQueued)
	newCodeRun.Params = map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	newCodeRun.Body = `{"action": "process"}`
	newCodeRun.Headers = map[string]interface{}{
		"Content-Type": "application/json",
		"User-Agent":   "TestClient/1.0",
	}

	createdRun, err := coderunRepo.Create(ctx, newCodeRun)
	if err != nil {
		return fmt.Errorf("failed to create coderun: %w", err)
	}

	log.Printf("Created coderun with ID: %s, Status: %s\n", createdRun.ID, createdRun.Status)

	// Update coderun to started
	createdRun.Status = coderun.StatusStarted
	updatedRun, err := coderunRepo.Update(ctx, createdRun.ID, createdRun)
	if err != nil {
		return fmt.Errorf("failed to update coderun: %w", err)
	}

	log.Printf("Updated coderun status to: %s\n", updatedRun.Status)

	// Complete the coderun with results
	updatedRun.Status = coderun.StatusCompleted
	updatedRun.Result = `{"success": true, "message": "Processing completed"}`
	updatedRun.Extra = map[string]interface{}{
		"status_code":  200,
		"content_type": "application/json",
	}

	finalRun, err := coderunRepo.Update(ctx, updatedRun.ID, updatedRun)
	if err != nil {
		return fmt.Errorf("failed to complete coderun: %w", err)
	}

	log.Printf("Completed coderun with result: %s\n", finalRun.Result)

	// Get coderun by ID
	retrievedRun, err := coderunRepo.GetByID(ctx, finalRun.ID)
	if err != nil {
		return fmt.Errorf("failed to get coderun: %w", err)
	}

	log.Printf("Retrieved coderun: ID=%s, Status=%s\n", retrievedRun.ID, retrievedRun.Status)

	// List coderuns by code ID
	filter := map[string]interface{}{}
	coderuns, err := coderunRepo.ListByCodeID(ctx, newCodeRun.CodeID, filter)
	if err != nil {
		return fmt.Errorf("failed to list coderuns: %w", err)
	}

	log.Printf("Found %d coderuns for code ID\n", len(coderuns))

	// List with date filters
	filterWithDates := map[string]interface{}{
		"after":  time.Now().Add(-24 * time.Hour),
		"before": time.Now(),
	}

	recentRuns, err := coderunRepo.ListByCodeID(ctx, newCodeRun.CodeID, filterWithDates)
	if err != nil {
		return fmt.Errorf("failed to list recent coderuns: %w", err)
	}

	log.Printf("Found %d coderuns in last 24 hours\n", len(recentRuns))

	// Delete old coderuns (cleanup)
	oldDate := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	deletedCount, err := coderunRepo.DeleteOlder(ctx, oldDate, 100)
	if err != nil {
		return fmt.Errorf("failed to delete old coderuns: %w", err)
	}

	log.Printf("Deleted %d old coderuns\n", deletedCount)

	return nil
}

// ExampleUsageWithService demonstrates typical usage pattern
func ExampleUsageWithService() error {
	cfg := &config.Config{
		DB: config.DBConfig{
			URI: "postgres://localhost:5432/codeactions?sslmode=disable",
		},
	}

	// Setup database
	db, err := postgre.GetPostgreDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	// Create repository
	repo := NewCodeRunRepository(db)
	ctx := context.Background()

	// Create and execute a coderun
	codeID := "507f1f77bcf86cd799439011"

	// 1. Queue execution
	run := coderun.NewCodeRun(codeID, coderun.StatusQueued)
	run.Params = map[string]interface{}{
		"input": "test data",
	}

	run, err = repo.Create(ctx, run)
	if err != nil {
		return fmt.Errorf("failed to queue run: %w", err)
	}

	// 2. Start execution
	run.Status = coderun.StatusStarted
	run, err = repo.Update(ctx, run.ID, run)
	if err != nil {
		return fmt.Errorf("failed to start run: %w", err)
	}

	// 3. Complete execution
	run.Status = coderun.StatusCompleted
	run.Result = `{"output": "result data"}`
	run.Extra = map[string]interface{}{
		"status_code":  200,
		"content_type": "application/json",
	}

	run, err = repo.Update(ctx, run.ID, run)
	if err != nil {
		return fmt.Errorf("failed to complete run: %w", err)
	}

	log.Printf("Execution completed: %s\n", run.Result)

	return nil
}
