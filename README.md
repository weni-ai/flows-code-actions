```
╔═╗┬  ┌─┐┬ ┬┌─┐  ╔═╗┌─┐┌┬┐┌─┐  ╔═╗┌─┐┌┬┐┬┌─┐┌┐┌┌─┐
╠╣ │  │ ││││└─┐  ║  │ │ ││├┤   ╠═╣│   │ ││ ││││└─┐
╚  ┴─┘└─┘└┴┘└─┘  ╚═╝└─┘─┴┘└─┘  ╩ ╩└─┘ ┴ ┴└─┘┘└┘└─┘
```

[Postman Collection](./doc/Code%20Actions.postman_collection_v3.json)

[WORK IN PROGRESS]

### requirements:

* go 1.21.4

* mongodb running local


### how to run

download dependencies:
```
go mod download -x
```

install air:
```
go install github.com/cosmtrek/air@latest
```

run:

```
air -d
```

### LocalStack S3 Development Setup

For local S3 development, you can use LocalStack to simulate Amazon S3:

#### 1. Start LocalStack
```bash
# Start LocalStack with S3 configured
./scripts/start-localstack.sh
```

#### 2. Configure environment variables
```bash
# Load LocalStack configuration
source config.localstack.example
```

#### 3. Test configuration
```bash
# Test basic S3 operations
./scripts/test-s3.sh
```

#### 4. Run application
```bash
# With S3 enabled via LocalStack
air -d
```

#### S3 environment variables:
- `FLOWS_CODE_ACTIONS_S3_ENABLED=true` - Enable S3 storage
- `FLOWS_CODE_ACTIONS_S3_ENDPOINT=http://localhost:4566` - LocalStack endpoint  
- `FLOWS_CODE_ACTIONS_S3_BUCKET_NAME=codeactions-dev` - Bucket name
- `FLOWS_CODE_ACTIONS_S3_REGION=us-east-1` - AWS region
- `FLOWS_CODE_ACTIONS_S3_PREFIX=codeactions` - Prefix for organization

#### Useful URLs:
- LocalStack Health: http://localhost:4566/health
- S3 Endpoint: http://localhost:4566

#### Available scripts:
- `./scripts/start-localstack.sh` - Start LocalStack (default mode)
- `./scripts/start-localstack-safe.sh` - Start LocalStack (alternative mode)
- `./scripts/test-s3-simple.sh` - Test basic S3 operations  
- `./scripts/cleanup-localstack.sh` - Completely cleanup LocalStack

#### Troubleshooting:
**"Device or resource busy" error:**
```bash
# Completely cleanup LocalStack
./scripts/cleanup-localstack.sh

# Then start again  
./scripts/start-localstack.sh
```

**LocalStack doesn't become "ready" after 30 attempts:**
- LocalStack is probably working, but the script didn't detect it
- Check manually: `curl http://localhost:4566/_localstack/health`
- Use alternative script: `./scripts/start-localstack-safe.sh`
- Run simple test: `./scripts/test-s3-simple.sh`
