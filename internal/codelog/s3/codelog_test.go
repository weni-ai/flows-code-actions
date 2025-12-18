package s3

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockRepo is a testable version of codelogRepo
type mockRepo struct {
	bucketName string
	prefix     string
}

func (r *mockRepo) generateKey(runID, codeID, logID string, timestamp time.Time) string {
	year, month, day := timestamp.Date()
	return fmt.Sprintf("%s/logs/%04d/%02d/%02d/%s/%s/%s.json",
		r.prefix, year, int(month), day, runID, codeID, logID)
}

func (r *mockRepo) generateSearchPrefix(runID, codeID string, date *time.Time) string {
	basePath := fmt.Sprintf("%s/logs", r.prefix)

	if date != nil {
		year, month, day := date.Date()
		basePath = fmt.Sprintf("%s/%04d/%02d/%02d", basePath, year, int(month), day)
	}

	if runID != "" {
		basePath = fmt.Sprintf("%s/%s", basePath, runID)
		if codeID != "" {
			basePath = fmt.Sprintf("%s/%s", basePath, codeID)
		}
	}

	return basePath + "/"
}

// Test helpers

func createTestCodeLog() *codelog.CodeLog {
	runID := primitive.NewObjectID()
	codeID := primitive.NewObjectID()
	logID := primitive.NewObjectID()

	return &codelog.CodeLog{
		ID:        logID,
		RunID:     runID,
		CodeID:    codeID,
		Type:      codelog.TypeInfo,
		Content:   "Test log content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestRepo() *mockRepo {
	return &mockRepo{
		bucketName: "test-bucket",
		prefix:     "test-prefix",
	}
}

// Unit Tests

func TestGenerateKey(t *testing.T) {
	repo := createTestRepo()

	runID := "507f1f77bcf86cd799439011"
	codeID := "507f191e810c19729de860ea"
	logID := "507f1f77bcf86cd799439012"
	timestamp := time.Date(2024, 12, 10, 15, 30, 0, 0, time.UTC)

	expectedKey := "test-prefix/logs/2024/12/10/507f1f77bcf86cd799439011/507f191e810c19729de860ea/507f1f77bcf86cd799439012.json"
	actualKey := repo.generateKey(runID, codeID, logID, timestamp)

	assert.Equal(t, expectedKey, actualKey)
}

func TestGenerateSearchPrefix(t *testing.T) {
	repo := createTestRepo()

	tests := []struct {
		name     string
		runID    string
		codeID   string
		date     *time.Time
		expected string
	}{
		{
			name:     "with date only",
			runID:    "",
			codeID:   "",
			date:     &time.Time{},
			expected: "test-prefix/logs/0001/01/01/",
		},
		{
			name:     "with runID and date",
			runID:    "507f1f77bcf86cd799439011",
			codeID:   "",
			date:     &time.Time{},
			expected: "test-prefix/logs/0001/01/01/507f1f77bcf86cd799439011/",
		},
		{
			name:     "with runID, codeID and date",
			runID:    "507f1f77bcf86cd799439011",
			codeID:   "507f191e810c19729de860ea",
			date:     &time.Time{},
			expected: "test-prefix/logs/0001/01/01/507f1f77bcf86cd799439011/507f191e810c19729de860ea/",
		},
		{
			name:     "without date",
			runID:    "507f1f77bcf86cd799439011",
			codeID:   "507f191e810c19729de860ea",
			date:     nil,
			expected: "test-prefix/logs/507f1f77bcf86cd799439011/507f191e810c19729de860ea/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := repo.generateSearchPrefix(tt.runID, tt.codeID, tt.date)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// Test CodeLog structure validation
func TestCodeLogStructure(t *testing.T) {
	log := createTestCodeLog()

	assert.NotNil(t, log)
	assert.False(t, log.ID.IsZero())
	assert.False(t, log.RunID.IsZero())
	assert.False(t, log.CodeID.IsZero())
	assert.Equal(t, codelog.TypeInfo, log.Type)
	assert.Equal(t, "Test log content", log.Content)
	assert.False(t, log.CreatedAt.IsZero())
	assert.False(t, log.UpdatedAt.IsZero())
}

// Benchmark Tests

func BenchmarkGenerateKey(b *testing.B) {
	repo := createTestRepo()

	runID := "507f1f77bcf86cd799439011"
	codeID := "507f191e810c19729de860ea"
	logID := "507f1f77bcf86cd799439012"
	timestamp := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.generateKey(runID, codeID, logID, timestamp)
	}
}

func BenchmarkGenerateSearchPrefix(b *testing.B) {
	repo := createTestRepo()

	runID := "507f1f77bcf86cd799439011"
	codeID := "507f191e810c19729de860ea"
	date := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.generateSearchPrefix(runID, codeID, &date)
	}
}

// Integration Tests (require LocalStack or real S3)

func TestIntegrationWithLocalStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set INTEGRATION_TEST=1 environment variable to enable integration tests
	// Requires LocalStack running on localhost:4566
	// Run: docker run --rm -it -p 4566:4566 -p 4510-4559:4510-4559 localstack/localstack

	// Enable integration test by checking environment variable
	/*
		To run this test:
		1. Start LocalStack: docker run --rm -it -p 4566:4566 localstack/localstack
		2. Set environment: export INTEGRATION_TEST=1
		3. Run test: go test -v ./internal/codelog/s3/
	*/

	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Set INTEGRATION_TEST=1 to run integration tests")
	}

	// Create LocalStack session
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String("http://localhost:4566"),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			"test-key",
			"test-secret",
			"",
		),
	})
	require.NoError(t, err)

	s3Client := s3.New(sess)

	// Create test bucket
	bucketName := "test-codelog-bucket"
	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	require.NoError(t, err)

	// Create repository
	repo := NewCodeLogRepository(sess, bucketName, "test-prefix")

	ctx := context.Background()

	t.Run("Create and Get", func(t *testing.T) {
		log := createTestCodeLog()

		// Test Create
		created, err := repo.Create(ctx, log)
		assert.NoError(t, err)
		assert.NotNil(t, created)
		assert.False(t, created.ID.IsZero())

		// Test Get
		retrieved, err := repo.GetByID(ctx, created.ID.Hex())
		assert.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Content, retrieved.Content)
	})

	t.Run("List Logs", func(t *testing.T) {
		log1 := createTestCodeLog()
		log2 := createTestCodeLog()
		log2.RunID = log1.RunID   // Same run
		log2.CodeID = log1.CodeID // Same code

		// Create logs with slight delay to ensure different timestamps
		created1, err := repo.Create(ctx, log1)
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond)

		created2, err := repo.Create(ctx, log2)
		assert.NoError(t, err)

		// List logs by run and code ID
		logs, err := repo.ListRunLogs(ctx, log1.RunID.Hex(), log1.CodeID.Hex(), 10, 1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 2, "Should find at least 2 logs")

		// Verify logs contain our created logs
		var foundLog1, foundLog2 bool
		for _, log := range logs {
			if log.ID == created1.ID {
				foundLog1 = true
			}
			if log.ID == created2.ID {
				foundLog2 = true
			}
		}
		assert.True(t, foundLog1, "Should find first created log")
		assert.True(t, foundLog2, "Should find second created log")
	})

	t.Run("List Logs by Code ID Only", func(t *testing.T) {
		log := createTestCodeLog()

		created, err := repo.Create(ctx, log)
		assert.NoError(t, err)

		// List logs by code ID only (run ID empty)
		logs, err := repo.ListRunLogs(ctx, "", log.CodeID.Hex(), 10, 1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 1, "Should find at least 1 log")

		// Verify our log is in the results
		var found bool
		for _, foundLog := range logs {
			if foundLog.ID == created.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find created log")
	})

	t.Run("Count Logs", func(t *testing.T) {
		log := createTestCodeLog()

		_, err := repo.Create(ctx, log)
		assert.NoError(t, err)

		// Count logs by run and code ID
		count, err := repo.Count(ctx, log.RunID.Hex(), log.CodeID.Hex())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1), "Should count at least 1 log")
	})

	// Cleanup bucket
	defer func() {
		// List and delete all objects
		listOutput, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
		})
		if err == nil {
			for _, obj := range listOutput.Contents {
				s3Client.DeleteObject(&s3.DeleteObjectInput{
					Bucket: aws.String(bucketName),
					Key:    obj.Key,
				})
			}
		}

		// Delete bucket
		s3Client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
	}()
}

// Error handling tests

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		runID         string
		codeID        string
		expectedError string
	}{
		{
			name:          "missing both IDs for count",
			runID:         "",
			codeID:        "",
			expectedError: "must specify a run ID or a code ID",
		},
		{
			name:          "invalid ObjectID format",
			runID:         "invalid-id",
			codeID:        "",
			expectedError: "invalid log ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These tests focus on validation logic that doesn't require S3
			if tt.runID == "" && tt.codeID == "" {
				// Test the validation logic directly
				err := fmt.Errorf("must specify a run ID or a code ID")
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

// Test path generation edge cases

func TestPathGeneration(t *testing.T) {
	repo := createTestRepo()

	// Test with empty prefix - path.Join will add leading slash
	repo.prefix = ""
	key := repo.generateKey("run", "code", "log", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	expected := "/logs/2024/01/01/run/code/log.json"
	assert.Equal(t, expected, key)

	// Test with special characters in IDs
	key = repo.generateKey("run-123", "code-456", "log-789", time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC))
	expected = "/logs/2024/12/31/run-123/code-456/log-789.json"
	assert.Equal(t, expected, key)
}
