package pg

import (
	"context"
	"fmt"
	"log"

	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	"github.com/weni-ai/flows-code-actions/internal/db/postgre"
)

// Example demonstrates how to use PostgreSQL codelib repository
func Example() error {
	// Example configuration
	cfg := &config.Config{
		DB: config.DBConfig{
			URI:                      "postgres://localhost:5432/codeactions?sslmode=disable",
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

	log.Println("________________-")

	// Connect to PostgreSQL
	db, err := postgre.GetPostgreDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	log.Println("Starting create codelib")

	// Create repository
	codelibRepo := NewCodeLibRepo(db)

	ctx := context.Background()

	// Create a new codelib entry
	newCodeLib := &codelib.CodeLib{
		Name:     "numpy",
		Language: codelib.TypePy,
	}

	createdCodeLib, err := codelibRepo.Create(ctx, newCodeLib)
	if err != nil {
		return fmt.Errorf("failed to create codelib: %w", err)
	}

	log.Printf("Created codelib with ID: %s\n", createdCodeLib.ID.Hex())

	// Create multiple codelibs in bulk
	bulkCodeLibs := []*codelib.CodeLib{
		{Name: "pandas", Language: codelib.TypePy},
		{Name: "requests", Language: codelib.TypePy},
		{Name: "matplotlib", Language: codelib.TypePy},
	}

	createdBulk, err := codelibRepo.CreateBulk(ctx, bulkCodeLibs)
	if err != nil {
		return fmt.Errorf("failed to create bulk codelibs: %w", err)
	}

	fmt.Printf("Created %d codelibs in bulk\n", len(createdBulk))

	// List all Python libraries
	pythonLang := codelib.TypePy
	allPythonLibs, err := codelibRepo.List(ctx, &pythonLang)
	if err != nil {
		return fmt.Errorf("failed to list Python libraries: %w", err)
	}

	fmt.Printf("Found %d Python libraries:\n", len(allPythonLibs))
	for _, lib := range allPythonLibs {
		fmt.Printf("- %s (ID: %s)\n", lib.Name, lib.ID.Hex())
	}

	// Find a specific library
	foundLib, err := codelibRepo.Find(ctx, "numpy", &pythonLang)
	if err != nil {
		return fmt.Errorf("failed to find numpy library: %w", err)
	}

	fmt.Printf("Found library: %s (Language: %s)\n", foundLib.Name, foundLib.Language)

	// List all libraries (no language filter)
	allLibs, err := codelibRepo.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list all libraries: %w", err)
	}

	fmt.Printf("Total libraries in database: %d\n", len(allLibs))

	return nil
}

// ExampleUsageWithService demonstrates how to use the repository with the service layer
func ExampleUsageWithService() error {
	cfg := &config.Config{
		DB: config.DBConfig{
			URI: "postgres://user:password@localhost:5432/dbname?sslmode=disable",
			// ... other config
		},
	}

	// Setup database connection
	db, err := postgre.GetPostgreDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer db.Close()

	// Create repository and service
	codelibRepo := NewCodeLibRepo(db)
	codelibService := codelib.NewCodeLibService(codelibRepo)

	ctx := context.Background()

	// Use service methods (same interface as MongoDB implementation)
	newLib := &codelib.CodeLib{
		Name:     "flask",
		Language: codelib.TypePy,
	}

	createdLib, err := codelibService.Create(ctx, newLib)
	if err != nil {
		return fmt.Errorf("failed to create library via service: %w", err)
	}

	fmt.Printf("Service created library: %s\n", createdLib.Name)

	// Find library via service
	pythonLang := codelib.TypePy
	foundLib, err := codelibService.Find(ctx, "flask", &pythonLang)
	if err != nil {
		return fmt.Errorf("failed to find library via service: %w", err)
	}

	fmt.Printf("Service found library: %s\n", foundLib.Name)

	return nil
}
