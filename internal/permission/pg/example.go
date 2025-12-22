package pg

import (
	"context"
	"fmt"

	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/db/postgre"
	"github.com/weni-ai/flows-code-actions/internal/permission"
	log "github.com/sirupsen/logrus"
)

// Example demonstrates how to use PostgreSQL user permission repository
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
	userRepo := NewUserRepository(db)

	ctx := context.Background()

	// Create a new user permission (Moderator)
	moderator := permission.NewUserPermission(
		"project-123-uuid",
		"moderator@example.com",
		permission.ModeratorRole,
	)

	_, err = userRepo.Create(ctx, moderator)
	if err != nil {
		return fmt.Errorf("failed to create moderator permission: %w", err)
	}

	log.Printf("Created moderator permission for %s\n", moderator.Email)

	// Create a viewer
	viewer := permission.NewUserPermission(
		"project-123-uuid",
		"viewer@example.com",
		permission.ViewerRole,
	)

	_, err = userRepo.Create(ctx, viewer)
	if err != nil {
		return fmt.Errorf("failed to create viewer permission: %w", err)
	}

	log.Printf("Created viewer permission for %s\n", viewer.Email)

	// Find user by email and project
	searchUser := &permission.UserPermission{
		Email:       "moderator@example.com",
		ProjectUUID: "project-123-uuid",
	}

	foundUser, err := userRepo.Find(ctx, searchUser)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	log.Printf("Found user: %s with role %d\n", foundUser.Email, foundUser.Role)

	// Check permissions
	if permission.HasPermission(foundUser, permission.WritePermission) {
		log.Printf("User %s has WRITE permission\n", foundUser.Email)
	}

	if permission.HasPermission(foundUser, permission.ReadPermission) {
		log.Printf("User %s has READ permission\n", foundUser.Email)
	}

	// Find by email only
	searchByEmail := &permission.UserPermission{
		Email: "viewer@example.com",
	}

	viewerUser, err := userRepo.Find(ctx, searchByEmail)
	if err != nil {
		return fmt.Errorf("failed to find viewer: %w", err)
	}

	log.Printf("Found viewer: %s (Role: %d)\n", viewerUser.Email, viewerUser.Role)

	// Update user role from Viewer to Contributor
	viewerUser.Role = permission.ContributorRole
	updatedUser, err := userRepo.Update(ctx, viewerUser.ID.Hex(), viewerUser)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	log.Printf("Updated user %s to role %d\n", updatedUser.Email, updatedUser.Role)

	// Try to create duplicate (should fail)
	duplicate := permission.NewUserPermission(
		"project-123-uuid",
		"moderator@example.com",
		permission.ModeratorRole,
	)

	_, err = userRepo.Create(ctx, duplicate)
	if err != nil {
		log.Printf("Expected error for duplicate: %v\n", err)
	}

	return nil
}

// ExamplePermissionChecks demonstrates permission checking
func ExamplePermissionChecks() error {
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

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create users with different roles
	roles := map[string]permission.Role{
		"viewer@example.com":      permission.ViewerRole,
		"contributor@example.com": permission.ContributorRole,
		"moderator@example.com":   permission.ModeratorRole,
	}

	projectUUID := "test-project-uuid"

	for email, role := range roles {
		user := permission.NewUserPermission(projectUUID, email, role)
		repo.Create(ctx, user)
	}

	// Check permissions for each user
	for email, role := range roles {
		searchUser := &permission.UserPermission{
			Email:       email,
			ProjectUUID: projectUUID,
		}

		user, err := repo.Find(ctx, searchUser)
		if err != nil {
			continue
		}

		readPerm := permission.HasPermission(user, permission.ReadPermission)
		writePerm := permission.HasPermission(user, permission.WritePermission)

		log.Printf("User: %s (Role: %d)\n", email, role)
		log.Printf("  - Read Permission: %v\n", readPerm)
		log.Printf("  - Write Permission: %v\n", writePerm)
	}

	return nil
}
