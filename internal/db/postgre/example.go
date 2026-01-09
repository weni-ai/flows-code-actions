package postgre

import (
	"context"
	"fmt"

	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/code"
	codeRepo "github.com/weni-ai/flows-code-actions/internal/code/pg"
)

// Example demonstrates how to use PostgreSQL database connection
func Example() error {
	// Example configuration (you would get this from your config system)
	cfg := &config.Config{
		DB: config.DBConfig{
			URI:                      "postgres://user:password@localhost:5432/dbname?sslmode=disable",
			Name:                     "codeactions",
			Timeout:                  30,
			MaxRetries:               3,
			MaxPoolSize:              25,
			MinPoolSize:              5,
			ConnectTimeoutSeconds:    10,
			ServerSelectionTimeout:   30,
			SocketTimeoutSeconds:     30,
			HeartbeatIntervalSeconds: 10,
		},
	}

	// Connect to PostgreSQL
	db, err := GetPostgreDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	// Perform health check
	healthResult := PerformHealthCheck(cfg)
	LogHealthStatus(healthResult)

	// Create repository
	codeRepository := codeRepo.NewCodeRepository(db)

	// Example usage
	ctx := context.Background()

	// Create a new code entry
	newCode := &code.Code{
		Name:        "example-code",
		Type:        code.TypeFlow,
		Source:      "print('Hello, World!')",
		Language:    code.TypePy,
		ProjectUUID: "project-123",
		Timeout:     60,
	}

	createdCode, err := codeRepository.Create(ctx, newCode)
	if err != nil {
		return fmt.Errorf("failed to create code: %w", err)
	}

	fmt.Printf("Created code with ID: %s\n", createdCode.ID)

	// Get code by ID
	retrievedCode, err := codeRepository.GetByID(ctx, createdCode.ID)
	if err != nil {
		return fmt.Errorf("failed to get code: %w", err)
	}

	fmt.Printf("Retrieved code: %s\n", retrievedCode.Name)

	// List codes by project
	codes, err := codeRepository.ListByProjectUUID(ctx, "project-123", "")
	if err != nil {
		return fmt.Errorf("failed to list codes: %w", err)
	}

	fmt.Printf("Found %d codes for project\n", len(codes))

	return nil
}

// ExamplePagination demonstrates how to use PostgreSQL pagination
func ExamplePagination() {
	// Create paginator for page 2 with 10 items per page
	paginator := NewPostgrePaginate(10, 2)

	// Get limit and offset
	limit, offset := paginator.GetLimitOffset()
	fmt.Printf("Limit: %d, Offset: %d\n", limit, offset)

	// Apply to a query
	baseQuery := "SELECT * FROM codes WHERE project_uuid = $1 ORDER BY created_at DESC"
	paginatedQuery := paginator.ApplyPagination(baseQuery)
	fmt.Printf("Paginated query: %s\n", paginatedQuery)
}
