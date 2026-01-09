package pg

import (
	"context"
	"fmt"

	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/db/postgre"
	"github.com/weni-ai/flows-code-actions/internal/project"
	log "github.com/sirupsen/logrus"
)

// Example demonstrates how to use PostgreSQL project repository
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
	projectRepo := NewProjectRepository(db)

	ctx := context.Background()

	// Create a new project
	newProject := project.NewProject(
		"project-uuid-123",
		"My Awesome Project",
	)

	// Add authorizations
	newProject.Authorizations = []struct {
		UserEmail string `json:"user_email"`
		Role      string `json:"role"`
	}{
		{
			UserEmail: "admin@example.com",
			Role:      "admin",
		},
		{
			UserEmail: "user@example.com",
			Role:      "member",
		},
	}

	createdProject, err := projectRepo.Create(ctx, newProject)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	log.Printf("Created project: %s (UUID: %s)\n", createdProject.Name, createdProject.UUID)
	log.Printf("Project ID: %s\n", createdProject.ID)
	log.Printf("Authorizations: %+v\n", createdProject.Authorizations)

	// Find project by UUID
	foundProject, err := projectRepo.FindByUUID(ctx, "project-uuid-123")
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	log.Printf("Found project: %s\n", foundProject.Name)
	log.Printf("Number of authorizations: %d\n", len(foundProject.Authorizations))

	// Update project name
	foundProject.Name = "My Updated Project Name"
	updatedProject, err := projectRepo.Update(ctx, foundProject)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	log.Printf("Updated project name to: %s\n", updatedProject.Name)

	// Update authorizations
	newAuthorizations := []struct {
		UserEmail string `json:"user_email"`
		Role      string `json:"role"`
	}{
		{
			UserEmail: "admin@example.com",
			Role:      "admin",
		},
		{
			UserEmail: "user@example.com",
			Role:      "member",
		},
		{
			UserEmail: "newuser@example.com",
			Role:      "viewer",
		},
	}

	err = projectRepo.UpdateAuthorizations(ctx, "project-uuid-123", newAuthorizations)
	if err != nil {
		return fmt.Errorf("failed to update authorizations: %w", err)
	}

	log.Println("Updated project authorizations")

	// Verify authorizations update
	verifyProject, err := projectRepo.FindByUUID(ctx, "project-uuid-123")
	if err != nil {
		return fmt.Errorf("failed to verify project: %w", err)
	}

	log.Printf("Verified project has %d authorizations\n", len(verifyProject.Authorizations))
	for i, auth := range verifyProject.Authorizations {
		log.Printf("  [%d] %s - %s\n", i+1, auth.UserEmail, auth.Role)
	}

	// Try to create duplicate (should fail)
	duplicate := project.NewProject("project-uuid-123", "Duplicate Project")
	_, err = projectRepo.Create(ctx, duplicate)
	if err != nil {
		log.Printf("Expected error for duplicate: %v\n", err)
	}

	return nil
}

// ExampleMultipleProjects demonstrates managing multiple projects
func ExampleMultipleProjects() error {
	cfg := &config.Config{
		DB: config.DBConfig{
			URI: "postgres://localhost:5432/codeactions?sslmode=disable",
		},
	}

	db, err := postgre.GetPostgreDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	repo := NewProjectRepository(db)
	ctx := context.Background()

	// Create multiple projects
	projects := []struct {
		UUID string
		Name string
	}{
		{"project-001", "E-commerce Platform"},
		{"project-002", "Mobile App Backend"},
		{"project-003", "Data Analytics Pipeline"},
	}

	for _, p := range projects {
		proj := project.NewProject(p.UUID, p.Name)
		proj.Authorizations = []struct {
			UserEmail string `json:"user_email"`
			Role      string `json:"role"`
		}{
			{
				UserEmail: "owner@example.com",
				Role:      "owner",
			},
		}

		createdProj, err := repo.Create(ctx, proj)
		if err != nil {
			log.Printf("Error creating project %s: %v\n", p.Name, err)
			continue
		}

		log.Printf("Created project: %s (UUID: %s)\n", createdProj.Name, createdProj.UUID)
	}

	// Retrieve and display all created projects
	for _, p := range projects {
		proj, err := repo.FindByUUID(ctx, p.UUID)
		if err != nil {
			continue
		}

		log.Printf("\nProject: %s\n", proj.Name)
		log.Printf("  UUID: %s\n", proj.UUID)
		log.Printf("  Created: %s\n", proj.CreatedAt.Format("2006-01-02 15:04:05"))
		log.Printf("  Authorizations: %d user(s)\n", len(proj.Authorizations))
	}

	return nil
}
