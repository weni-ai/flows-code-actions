# CodeLog Migration to S3

This document describes how to migrate CodeLog log storage from MongoDB to Amazon S3.

## Overview

The S3 implementation for CodeLog offers:

- **Scalable storage**: Unlimited storage capacity
- **Reduced costs**: Cheaper storage for historical logs
- **Durability**: 99.999999999% (11 9's) durability
- **Temporal organization**: Logs organized by date/run/code for efficient searching
- **Compatibility**: Maintains the same interface as the MongoDB implementation

## S3 Storage Structure

Logs are organized in the following structure:

```
{bucket}/{prefix}/logs/{year}/{month}/{day}/{run_id}/{code_id}/{log_id}.json
```

Example:
```
my-bucket/codeactions/logs/2024/12/10/507f1f77bcf86cd799439011/507f191e810c19729de860ea/507f1f77bcf86cd799439012.json
```

## Configuration

### Environment Variables

```bash
# Enable S3 storage for CodeLog
FLOWS_CODE_ACTIONS_S3_ENABLED=true

# AWS S3 settings
FLOWS_CODE_ACTIONS_S3_REGION=us-east-1
FLOWS_CODE_ACTIONS_S3_BUCKET_NAME=my-codeactions-logs
FLOWS_CODE_ACTIONS_S3_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
FLOWS_CODE_ACTIONS_S3_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
FLOWS_CODE_ACTIONS_S3_PREFIX=codeactions

# Maintain MongoDB settings for other data
FLOWS_CODE_ACTIONS_MONGO_DB_URI=mongodb://localhost:27017
FLOWS_CODE_ACTIONS_MONGO_DB_NAME=code-actions
```

### S3 Bucket Configuration

1. **Create S3 bucket**:
```bash
aws s3 mb s3://my-codeactions-logs --region us-east-1
```

2. **Configure lifecycle policy** (optional):
```json
{
    "Rules": [
        {
            "ID": "CodeLogLifecycle",
            "Status": "Enabled",
            "Filter": {
                "Prefix": "codeactions/logs/"
            },
            "Transitions": [
                {
                    "Days": 30,
                    "StorageClass": "STANDARD_IA"
                },
                {
                    "Days": 90,
                    "StorageClass": "GLACIER"
                },
                {
                    "Days": 365,
                    "StorageClass": "DEEP_ARCHIVE"
                }
            ]
        }
    ]
}
```

3. **Configure CORS** (if needed for web access):
```json
[
    {
        "AllowedHeaders": ["*"],
        "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
        "AllowedOrigins": ["*"],
        "ExposeHeaders": []
    }
]
```

## Migration Strategies

### Option 1: Immediate Migration (Big Bang)

1. **Stop application**
2. **Migrate existing data** (migration script)
3. **Update configuration** to use S3
4. **Restart application**

**Pros**: Quick and clean migration
**Cons**: Downtime required

### Option 2: Gradual Migration (Dual Write)

1. **Implement dual write**: Write to both MongoDB and S3 simultaneously
2. **Migrate historical data** in background
3. **Validate consistency** between both systems
4. **Disable MongoDB** after complete validation

**Pros**: Zero downtime
**Cons**: Higher complexity and temporary resource usage

### Option 3: Migration by Date (Recommended)

1. **Configure S3** for new logs
2. **Keep MongoDB** for historical logs (read-only)
3. **Implement fallback** to search in both systems
4. **Migrate old data** gradually or archive

**Pros**: Low risk, flexible
**Cons**: Search complexity across multiple systems

## Migration Script

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/weni-ai/flows-code-actions/config"
    "github.com/weni-ai/flows-code-actions/internal/codelog"
    "github.com/weni-ai/flows-code-actions/internal/codelog/mongodb"
    s3repo "github.com/weni-ai/flows-code-actions/internal/codelog/s3"
    "github.com/weni-ai/flows-code-actions/internal/db"
)

func main() {
    cfg := config.NewConfig()
    
    // Connect to MongoDB
    mongoDb, err := db.GetMongoDatabase(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    mongoRepo := mongodb.NewCodeLogRepository(mongoDb)
    
    // Connect to S3
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(cfg.S3.Region),
    })
    if err != nil {
        log.Fatal(err)
    }
    
    s3Repo := s3repo.NewCodeLogRepository(sess, cfg.S3.BucketName, cfg.S3.Prefix)
    
    // Migrate logs in batches
    ctx := context.Background()
    batchSize := 1000
    
    // Implement migration logic here
    // 1. Fetch logs from MongoDB in batches
    // 2. Convert and save to S3
    // 3. Verify integrity
    // 4. Mark as migrated or delete from MongoDB
    
    fmt.Println("Migration completed!")
}
```

## Monitoring and Observability

### Important Metrics

1. **S3 operations latency**
2. **Operation error rate**
3. **Storage and transfer costs**
4. **Search performance**

### Audit Logs

```go
// Structured logging example
logrus.WithFields(logrus.Fields{
    "operation": "codelog_create",
    "storage":   "s3",
    "bucket":    bucketName,
    "key":       key,
    "duration":  duration,
}).Info("CodeLog created in S3")
```

## Performance Considerations

### Optimizations

1. **Date-based search**: Limit search to last 7-30 days
2. **Cache**: Implement Redis cache for frequently accessed logs
3. **Pagination**: Use efficient pagination for large lists
4. **Compression**: Compress JSON logs before upload

### Limitations

1. **Search by ID**: Less efficient than MongoDB (requires date-based search)
2. **Transactions**: S3 doesn't support ACID transactions
3. **Consistency**: Eventual consistency in some regions
4. **Latency**: May be higher than local MongoDB

## Rollback

To revert to MongoDB:

1. **Change configuration**: `FLOWS_CODE_ACTIONS_S3_ENABLED=false`
2. **Restart application**
3. **Migrate critical data** back from S3 (if needed)

## Estimated Costs

### S3 Standard (us-east-1)
- **Storage**: $0.023 per GB/month
- **PUT Requests**: $0.0005 per 1,000 requests
- **GET Requests**: $0.0004 per 1,000 requests

### Example Calculation
- 1M logs/month × 2KB/log = 2GB/month
- Storage cost: 2GB × $0.023 = $0.046/month
- Request costs: Very low for typical usage

## Security

### Best Practices

1. **IAM Roles**: Use roles instead of access keys when possible
2. **Bucket Policies**: Restrict bucket access
3. **Encryption**: Enable encryption at rest and in transit
4. **VPC Endpoints**: Use VPC endpoints for private traffic

### Example IAM Policy

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject",
                "s3:ListBucket"
            ],
            "Resource": [
                "arn:aws:s3:::my-codeactions-logs",
                "arn:aws:s3:::my-codeactions-logs/*"
            ]
        }
    ]
}
```

## Testing

### Integration Tests

```go
func TestS3CodeLogRepository(t *testing.T) {
    // Setup test S3 bucket
    // Test CRUD operations
    // Verify data integrity
    // Test error scenarios
}
```

### Performance Tests

```go
func BenchmarkS3Operations(b *testing.B) {
    // Benchmark create, read, list operations
    // Compare with MongoDB performance
}
```

## Conclusion

Migration to S3 offers significant benefits in terms of cost and scalability, especially for historical logs. The implementation maintains compatibility with the existing interface, facilitating gradual migration.

We recommend starting with **Option 3 (Migration by Date)** to minimize risks and allow easy rollback if needed.