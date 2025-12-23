#!/bin/bash

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
while ! curl -s http://localhost:4566/_localstack/health | grep -q '"s3": "available"'; do
    echo "Waiting for S3 service..."
    sleep 2
done

echo "Creating S3 bucket for codeactions..."

# Create the bucket
aws --endpoint-url=http://localhost:4566 s3 mb s3://codeactions-dev --region us-east-1

# Enable versioning (optional)
aws --endpoint-url=http://localhost:4566 s3api put-bucket-versioning \
    --bucket codeactions-dev \
    --versioning-configuration Status=Enabled

# Create test folder structure (create empty object to simulate folder)
aws --endpoint-url=http://localhost:4566 s3api put-object \
    --bucket codeactions-dev \
    --key codeactions/logs/.gitkeep \
    --body /dev/null 2>/dev/null || true

echo "S3 bucket 'codeactions-dev' created successfully!"

# List buckets to confirm
echo "Available buckets:"
aws --endpoint-url=http://localhost:4566 s3 ls

