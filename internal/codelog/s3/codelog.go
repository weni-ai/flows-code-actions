package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type codelogRepo struct {
	s3Client   *s3.S3
	bucketName string
	prefix     string
}

// NewCodeLogRepository creates a new S3-based CodeLog repository
func NewCodeLogRepository(sess *session.Session, bucketName, prefix string) codelog.Repository {
	return &codelogRepo{
		s3Client:   s3.New(sess),
		bucketName: bucketName,
		prefix:     prefix,
	}
}

// generateKey creates S3 key for a log entry
// Format: {prefix}/logs/{year}/{month}/{day}/{run_id}/{code_id}/{log_id}.json
func (r *codelogRepo) generateKey(runID, codeID, logID string, timestamp time.Time) string {
	year, month, day := timestamp.Date()
	return path.Join(
		r.prefix,
		"logs",
		fmt.Sprintf("%04d", year),
		fmt.Sprintf("%02d", int(month)),
		fmt.Sprintf("%02d", day),
		runID,
		codeID,
		fmt.Sprintf("%s.json", logID),
	)
}

// generateSearchPrefix creates prefix for searching logs
func (r *codelogRepo) generateSearchPrefix(runID, codeID string, date *time.Time) string {
	basePath := path.Join(r.prefix, "logs")

	if date != nil {
		year, month, day := date.Date()
		basePath = path.Join(basePath,
			fmt.Sprintf("%04d", year),
			fmt.Sprintf("%02d", int(month)),
			fmt.Sprintf("%02d", day),
		)
	}

	if runID != "" {
		basePath = path.Join(basePath, runID)
		if codeID != "" {
			basePath = path.Join(basePath, codeID)
		}
	}

	return basePath + "/"
}

// generateSearchPrefixes creates multiple prefixes for searching logs
// Returns two prefixes: one with potential bucket name duplication (legacy) and one correct
// This handles the case where logs were saved with bucket name in the path due to endpoint misconfiguration
func (r *codelogRepo) generateSearchPrefixes(runID, codeID string, date *time.Time) []string {
	prefixes := make([]string, 0, 2)

	// First try: Legacy path with bucket name duplication (if bucket name is in prefix)
	// This handles logs saved when endpoint included bucket name
	// e.g., "weni-staging-codeactions/codeactions/logs/..."
	if strings.Contains(r.bucketName, "-") {
		legacyBasePath := path.Join(r.bucketName, r.prefix, "logs")

		if date != nil {
			year, month, day := date.Date()
			legacyBasePath = path.Join(legacyBasePath,
				fmt.Sprintf("%04d", year),
				fmt.Sprintf("%02d", int(month)),
				fmt.Sprintf("%02d", day),
			)
		}

		if runID != "" {
			legacyBasePath = path.Join(legacyBasePath, runID)
			if codeID != "" {
				legacyBasePath = path.Join(legacyBasePath, codeID)
			}
		}

		prefixes = append(prefixes, legacyBasePath+"/")
	}

	// Second try: Correct path (current implementation)
	prefixes = append(prefixes, r.generateSearchPrefix(runID, codeID, date))

	return prefixes
}

func (r *codelogRepo) Create(ctx context.Context, log *codelog.CodeLog) (*codelog.CodeLog, error) {
	// Generate new ID if not present
	if log.ID == "" {
		log.ID = primitive.NewObjectID().Hex()
	}

	// Set timestamps
	now := time.Now()
	log.CreatedAt = now
	log.UpdatedAt = now

	// Serialize log to JSON
	logData, err := json.Marshal(log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal log to JSON")
	}

	// Generate S3 key - IDs are now strings (UUIDs or ObjectID hex)
	key := r.generateKey(log.RunID, log.CodeID, log.ID, now)

	// Upload to S3
	_, err = r.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(logData),
		ContentType: aws.String("application/json"),
		Metadata: map[string]*string{
			"run-id":     aws.String(log.RunID),
			"code-id":    aws.String(log.CodeID),
			"log-type":   aws.String(string(log.Type)),
			"created-at": aws.String(now.Format(time.RFC3339)),
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload log to S3")
	}

	return log, nil
}

func (r *codelogRepo) GetByID(ctx context.Context, id string) (*codelog.CodeLog, error) {
	// Since we don't know the exact path, we need to search for the log
	// This is less efficient than MongoDB but still workable
	if id == "" {
		return nil, errors.New("invalid log ID: empty string")
	}

	// Search across recent dates (last 30 days) for the log
	// Use UTC to match Python's timezone when saving logs
	now := time.Now().UTC()
	for i := 0; i < 30; i++ {
		searchDate := now.AddDate(0, 0, -i)
		prefixes := r.generateSearchPrefixes("", "", &searchDate)

		// Try all possible prefixes (legacy and correct)
		for _, prefix := range prefixes {
			log, err := r.searchLogByID(ctx, prefix, id)
			if err != nil {
				continue // Try next prefix
			}
			if log != nil {
				return log, nil
			}
		}
	}

	return nil, errors.New("log not found")
}

func (r *codelogRepo) searchLogByID(ctx context.Context, prefix, logID string) (*codelog.CodeLog, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucketName),
		Prefix: aws.String(prefix),
	}

	var foundKey string
	err := r.s3Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			if strings.Contains(*obj.Key, logID+".json") {
				foundKey = *obj.Key // Store the actual key found
				return false        // Found it, stop pagination
			}
		}
		return true // Continue pagination
	})

	if err != nil {
		return nil, err
	}

	// If we found the object, retrieve it using the actual key
	if foundKey == "" {
		return nil, nil // Not found in this prefix
	}

	return r.getLogFromS3(ctx, foundKey)
}

func (r *codelogRepo) getLogFromS3(ctx context.Context, key string) (*codelog.CodeLog, error) {
	result, err := r.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	var log codelog.CodeLog
	if err := json.NewDecoder(result.Body).Decode(&log); err != nil {
		return nil, errors.Wrap(err, "failed to decode log from JSON")
	}

	return &log, nil
}

func (r *codelogRepo) ListRunLogs(ctx context.Context, runID, codeID string, limit, page int) ([]codelog.CodeLog, error) {
	if runID == "" && codeID == "" {
		return nil, errors.New("must specify a run ID or a code ID")
	}

	var logs []codelog.CodeLog

	// Search across recent dates (last 7 days for performance)
	// Use UTC to match Python's timezone when saving logs
	now := time.Now().UTC()
	for i := 0; i < 7; i++ {
		searchDate := now.AddDate(0, 0, -i)
		prefixes := r.generateSearchPrefixes(runID, codeID, &searchDate)

		// Try all possible prefixes (legacy with bucket name duplication and correct)
		for _, prefix := range prefixes {
			dayLogs, err := r.listLogsFromPrefix(ctx, prefix, runID, codeID)
			if err != nil {
				continue // Skip this prefix on error
			}

			logs = append(logs, dayLogs...)
		}
	}

	// Sort logs by creation time (newest first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})

	// Apply pagination
	start := (page - 1) * limit
	if start >= len(logs) {
		return []codelog.CodeLog{}, nil
	}

	end := start + limit
	if end > len(logs) {
		end = len(logs)
	}

	return logs[start:end], nil
}

func (r *codelogRepo) listLogsFromPrefix(ctx context.Context, prefix, runID, codeID string) ([]codelog.CodeLog, error) {
	var logs []codelog.CodeLog

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucketName),
		Prefix: aws.String(prefix),
	}

	err := r.s3Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			// Skip if not a JSON file
			if !strings.HasSuffix(*obj.Key, ".json") {
				continue
			}

			// Filter by runID/codeID if specified and not in path
			if runID != "" && !strings.Contains(*obj.Key, runID) {
				continue
			}
			if codeID != "" && !strings.Contains(*obj.Key, codeID) {
				continue
			}

			log, err := r.getLogFromS3(ctx, *obj.Key)
			if err != nil {
				continue // Skip invalid logs
			}

			// Double-check filtering (in case path structure doesn't match)
			if runID != "" && log.RunID != runID {
				continue
			}
			if codeID != "" && log.CodeID != codeID {
				continue
			}

			logs = append(logs, *log)
		}
		return true // Continue pagination
	})

	return logs, err
}

func (r *codelogRepo) Update(ctx context.Context, id, content string) (*codelog.CodeLog, error) {
	// First, get the existing log
	log, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update content and timestamp
	log.Content = content
	log.UpdatedAt = time.Now()

	// Delete old version and create new one
	// Note: This is not atomic, but S3 doesn't support atomic updates
	oldKey := r.generateKey(log.RunID, log.CodeID, log.ID, log.CreatedAt)

	// Upload updated version
	logData, err := json.Marshal(log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal updated log")
	}

	newKey := r.generateKey(log.RunID, log.CodeID, log.ID, log.UpdatedAt)

	_, err = r.s3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(newKey),
		Body:        bytes.NewReader(logData),
		ContentType: aws.String("application/json"),
		Metadata: map[string]*string{
			"run-id":     aws.String(log.RunID),
			"code-id":    aws.String(log.CodeID),
			"log-type":   aws.String(string(log.Type)),
			"updated-at": aws.String(log.UpdatedAt.Format(time.RFC3339)),
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload updated log")
	}

	// Delete old version if key changed
	if oldKey != newKey {
		_, _ = r.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(r.bucketName),
			Key:    aws.String(oldKey),
		})
	}

	return log, nil
}

func (r *codelogRepo) Delete(ctx context.Context, id string) error {
	// Find and delete the log
	log, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	key := r.generateKey(log.RunID, log.CodeID, log.ID, log.CreatedAt)

	_, err = r.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})

	return err
}

func (r *codelogRepo) DeleteOlder(ctx context.Context, date time.Time, limit int64) (int64, error) {
	var deletedCount int64

	// List objects older than the specified date
	// We'll search day by day backwards from the date
	searchDate := date
	for deletedCount < limit {
		prefixes := r.generateSearchPrefixes("", "", &searchDate)

		// Try all possible prefixes (legacy and correct)
		for _, prefix := range prefixes {
			input := &s3.ListObjectsV2Input{
				Bucket:  aws.String(r.bucketName),
				Prefix:  aws.String(prefix),
				MaxKeys: aws.Int64(1000), // Process in batches
			}

			result, err := r.s3Client.ListObjectsV2WithContext(ctx, input)
			if err != nil {
				continue // Skip this prefix on error
			}

			// Delete objects from this date
			for _, obj := range result.Contents {
				if deletedCount >= limit {
					break
				}

				// Check if object is actually older than date
				if obj.LastModified.After(date) {
					continue
				}

				_, err := r.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String(r.bucketName),
					Key:    obj.Key,
				})

				if err == nil {
					deletedCount++
				}
			}

			if deletedCount >= limit {
				break
			}
		}

		// Move to previous day
		searchDate = searchDate.AddDate(0, 0, -1)

		// Stop if we've gone back too far (e.g., 1 year)
		if time.Since(searchDate) > 365*24*time.Hour {
			break
		}
	}

	return deletedCount, nil
}

func (r *codelogRepo) Count(ctx context.Context, runID, codeID string) (int64, error) {
	if runID == "" && codeID == "" {
		return 0, errors.New("must specify a run ID or a code ID")
	}

	var count int64

	// Count across recent dates (last 30 days)
	// Use UTC to match Python's timezone when saving logs
	now := time.Now().UTC()
	for i := 0; i < 30; i++ {
		searchDate := now.AddDate(0, 0, -i)
		prefixes := r.generateSearchPrefixes(runID, codeID, &searchDate)

		// Try all possible prefixes (legacy and correct)
		for _, prefix := range prefixes {
			input := &s3.ListObjectsV2Input{
				Bucket: aws.String(r.bucketName),
				Prefix: aws.String(prefix),
			}

			err := r.s3Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
				for _, obj := range page.Contents {
					// Skip if not a JSON file
					if !strings.HasSuffix(*obj.Key, ".json") {
						continue
					}

					// Filter by runID/codeID if specified and not in path
					if runID != "" && !strings.Contains(*obj.Key, runID) {
						continue
					}
					if codeID != "" && !strings.Contains(*obj.Key, codeID) {
						continue
					}

					count++
				}
				return true
			})

			if err != nil {
				continue // Skip this prefix on error
			}
		}
	}

	return count, nil
}
