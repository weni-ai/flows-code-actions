# 8. Configuração

## Visão Geral

O Flows Code Actions é configurado inteiramente via variáveis de ambiente, seguindo os princípios do [12-Factor App](https://12factor.net/).

## Variáveis de Ambiente

### HTTP Server

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_HOST` | `:` | Host do servidor |
| `FLOWS_CODE_ACTIONS_PORT` | `8050` | Porta do servidor |

### Database

#### Configuração Geral

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_DB_TYPE` | `postgres` | Tipo do banco (`postgres` ou `mongodb`) |

#### PostgreSQL

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_DB_URI` | `postgres://test:test@localhost:5432/codeactions?sslmode=disable` | URI de conexão |
| `FLOWS_CODE_ACTIONS_DB_NAME` | `codeactions` | Nome do banco |
| `FLOWS_CODE_ACTIONS_POSTGRES_DB_TIMEOUT` | `35` | Timeout de conexão (segundos) |
| `FLOWS_CODE_ACTIONS_DB_MAX_RETRIES` | `3` | Máximo de tentativas de conexão |
| `FLOWS_CODE_ACTIONS_DB_MAX_POOL_SIZE` | `100` | Máximo de conexões no pool |
| `FLOWS_CODE_ACTIONS_DB_MIN_POOL_SIZE` | `5` | Mínimo de conexões no pool |
| `FLOWS_CODE_ACTIONS_DB_CONNECT_TIMEOUT` | `30` | Timeout de conexão (segundos) |
| `FLOWS_CODE_ACTIONS_DB_SOCKET_TIMEOUT` | `30` | Timeout de socket (segundos) |

#### MongoDB (Legado)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_MONGO_DB_URI` | `mongodb://localhost:27017` | URI de conexão |
| `FLOWS_CODE_ACTIONS_MONGO_DB_NAME` | `code-actions` | Nome do banco |
| `FLOWS_CODE_ACTIONS_MONGO_DB_TIMEOUT` | `35` | Timeout (segundos) |
| `FLOWS_CODE_ACTIONS_MONGO_DB_MAX_RETRIES` | `3` | Máximo de tentativas |
| `FLOWS_CODE_ACTIONS_MONGO_DB_MAX_POOL_SIZE` | `100` | Máximo de conexões |
| `FLOWS_CODE_ACTIONS_MONGO_DB_MIN_POOL_SIZE` | `5` | Mínimo de conexões |

### Redis

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_REDIS` | `redis://localhost:6379/10` | URI de conexão Redis |

### Autenticação

#### Auth Token (Modo Simples)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_AUTH_TOKEN` | `` | Token de autenticação |

#### OIDC (OpenID Connect)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_OIDC_AUTH_ENABLED` | `false` | Habilitar autenticação OIDC |
| `FLOWS_CODE_ACTIONS_OIDC_REALM` | `` | Realm do OIDC |
| `FLOWS_CODE_ACTIONS_OIDC_HOST` | `` | Host do provedor OIDC |

### RabbitMQ (Event-Driven)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_RABBITMQ_URL` | `` | URL de conexão RabbitMQ |
| `FLOWS_CODE_ACTIONS_PROJECT_EXCHANGE` | `` | Nome do exchange de projetos |
| `FLOWS_CODE_ACTIONS_PROJECT_QUEUE` | `` | Nome da queue de projetos |
| `FLOWS_CODE_ACTIONS_PERMISSION_EXCHANGE` | `` | Nome do exchange de permissões |
| `FLOW_CODE_ACTIONS_PERMISSION_QUEUE` | `` | Nome da queue de permissões |

### Amazon S3 (Armazenamento de Logs)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_S3_ENABLED` | `true` | Habilitar armazenamento S3 |
| `FLOWS_CODE_ACTIONS_S3_REGION` | `us-east-1` | Região AWS |
| `FLOWS_CODE_ACTIONS_S3_BUCKET_NAME` | `codeactions-dev` | Nome do bucket |
| `FLOWS_CODE_ACTIONS_S3_ACCESS_KEY_ID` | `test` | Access Key ID |
| `FLOWS_CODE_ACTIONS_S3_SECRET_ACCESS_KEY` | `test` | Secret Access Key |
| `FLOWS_CODE_ACTIONS_S3_PREFIX` | `codeactions` | Prefixo para objetos |
| `FLOWS_CODE_ACTIONS_S3_ENDPOINT` | `` | Endpoint customizado (LocalStack) |

### Rate Limiter

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_CODE_LIMITER_MAX_REQUESTS` | `600` | Máximo de requisições |
| `FLOWS_CODE_ACTIONS_CODE_LIMITER_WINDOW_WINDOW` | `60` | Janela de tempo (segundos) |

### Cleaner (Limpeza de Dados)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_CLEANER_SCHEDULE_TIME` | `01:00` | Horário de execução |
| `FLOWS_CODE_ACTIONS_CLEANER_RETENTION_PERIOD` | `30` | Período de retenção (dias) |

### Resource Management (CGroups)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_RESOURCE_ENABLED` | `false` | Habilitar gerenciamento de recursos |
| `FLOWS_CODE_ACTIONS_CPU_SHARES` | `0` | CPU shares |
| `FLOWS_CODE_ACTIONS_CPU_QUOTA` | `0` | CPU quota |
| `FLOWS_CODE_ACTIONS_MEMORY_LIMIT` | `0` | Limite de memória |
| `FLOWS_CODE_ACTIONS_MEMORY_RESERVATION` | `0` | Reserva de memória |

### Logging e Monitoramento

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_LOG_LEVEL` | `debug` | Nível de log (debug, info, warn, error) |
| `FLOWS_CODE_ACTIONS_SENTRY_DSN` | `` | DSN do Sentry para erros |
| `FLOWS_CODE_ACTIONS_ENVIRONMENT` | `local` | Ambiente (local, staging, production) |

### Segurança

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_BLACKLIST` | `` | Termos proibidos (separados por vírgula) |
| `FLOWS_CODE_ACTIONS_SKIPLIST` | `` | Termos a ignorar (separados por vírgula) |

### Health Check

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `FLOWS_CODE_ACTIONS_HEALTH_CHECK_CACHE_TIME` | `3` | Tempo de cache do health check |

---

## Exemplos de Configuração

### Desenvolvimento Local

```bash
# .env.local
FLOWS_CODE_ACTIONS_PORT=8050
FLOWS_CODE_ACTIONS_DB_TYPE=postgres
FLOWS_CODE_ACTIONS_DB_URI=postgres://test:test@localhost:5432/codeactions?sslmode=disable
FLOWS_CODE_ACTIONS_REDIS=redis://localhost:6379/10
FLOWS_CODE_ACTIONS_AUTH_TOKEN=dev-token-123
FLOWS_CODE_ACTIONS_LOG_LEVEL=debug
FLOWS_CODE_ACTIONS_ENVIRONMENT=local

# S3 com LocalStack
FLOWS_CODE_ACTIONS_S3_ENABLED=true
FLOWS_CODE_ACTIONS_S3_ENDPOINT=http://localhost:4566
FLOWS_CODE_ACTIONS_S3_BUCKET_NAME=codeactions-dev
FLOWS_CODE_ACTIONS_S3_REGION=us-east-1
```

### Staging

```bash
# .env.staging
FLOWS_CODE_ACTIONS_PORT=8050
FLOWS_CODE_ACTIONS_DB_TYPE=postgres
FLOWS_CODE_ACTIONS_DB_URI=postgres://user:pass@db-staging:5432/codeactions?sslmode=require
FLOWS_CODE_ACTIONS_REDIS=redis://redis-staging:6379/10
FLOWS_CODE_ACTIONS_ENVIRONMENT=staging
FLOWS_CODE_ACTIONS_LOG_LEVEL=info

# OIDC
FLOWS_CODE_ACTIONS_OIDC_AUTH_ENABLED=true
FLOWS_CODE_ACTIONS_OIDC_REALM=staging-realm
FLOWS_CODE_ACTIONS_OIDC_HOST=https://auth-staging.example.com

# RabbitMQ
FLOWS_CODE_ACTIONS_RABBITMQ_URL=amqp://user:pass@rabbitmq-staging:5672
FLOWS_CODE_ACTIONS_PROJECT_EXCHANGE=projects
FLOWS_CODE_ACTIONS_PROJECT_QUEUE=codeactions-projects
FLOWS_CODE_ACTIONS_PERMISSION_EXCHANGE=permissions
FLOW_CODE_ACTIONS_PERMISSION_QUEUE=codeactions-permissions

# S3
FLOWS_CODE_ACTIONS_S3_ENABLED=true
FLOWS_CODE_ACTIONS_S3_BUCKET_NAME=codeactions-staging
FLOWS_CODE_ACTIONS_S3_REGION=us-east-1

# Sentry
FLOWS_CODE_ACTIONS_SENTRY_DSN=https://xxx@sentry.io/123
```

### Produção

```bash
# .env.production
FLOWS_CODE_ACTIONS_PORT=8050
FLOWS_CODE_ACTIONS_DB_TYPE=postgres
FLOWS_CODE_ACTIONS_DB_URI=postgres://user:pass@db-prod:5432/codeactions?sslmode=require
FLOWS_CODE_ACTIONS_DB_MAX_POOL_SIZE=200
FLOWS_CODE_ACTIONS_DB_MIN_POOL_SIZE=20
FLOWS_CODE_ACTIONS_REDIS=redis://redis-prod:6379/10
FLOWS_CODE_ACTIONS_ENVIRONMENT=production
FLOWS_CODE_ACTIONS_LOG_LEVEL=warn

# OIDC
FLOWS_CODE_ACTIONS_OIDC_AUTH_ENABLED=true
FLOWS_CODE_ACTIONS_OIDC_REALM=production-realm
FLOWS_CODE_ACTIONS_OIDC_HOST=https://auth.example.com

# RabbitMQ
FLOWS_CODE_ACTIONS_RABBITMQ_URL=amqp://user:pass@rabbitmq-prod:5672
FLOWS_CODE_ACTIONS_PROJECT_EXCHANGE=projects
FLOWS_CODE_ACTIONS_PROJECT_QUEUE=codeactions-projects
FLOWS_CODE_ACTIONS_PERMISSION_EXCHANGE=permissions
FLOW_CODE_ACTIONS_PERMISSION_QUEUE=codeactions-permissions

# S3
FLOWS_CODE_ACTIONS_S3_ENABLED=true
FLOWS_CODE_ACTIONS_S3_BUCKET_NAME=codeactions-prod
FLOWS_CODE_ACTIONS_S3_REGION=us-east-1

# Rate Limiter
FLOWS_CODE_ACTIONS_CODE_LIMITER_MAX_REQUESTS=1000
FLOWS_CODE_ACTIONS_CODE_LIMITER_WINDOW_WINDOW=60

# Cleaner
FLOWS_CODE_ACTIONS_CLEANER_SCHEDULE_TIME=03:00
FLOWS_CODE_ACTIONS_CLEANER_RETENTION_PERIOD=90

# Resource Management
FLOWS_CODE_ACTIONS_RESOURCE_ENABLED=true
FLOWS_CODE_ACTIONS_CPU_QUOTA=50000
FLOWS_CODE_ACTIONS_MEMORY_LIMIT=536870912

# Sentry
FLOWS_CODE_ACTIONS_SENTRY_DSN=https://xxx@sentry.io/456

# Security
FLOWS_CODE_ACTIONS_BLACKLIST=os.system,subprocess,eval,exec
```

---

## Diagrama de Dependências

```mermaid
flowchart TB
    subgraph "Obrigatórios"
        A[PostgreSQL/MongoDB]
        B[Redis]
    end
    
    subgraph "Opcionais"
        C[RabbitMQ]
        D[S3/LocalStack]
        E[OIDC Provider]
        F[Sentry]
    end
    
    subgraph "Code Actions"
        G[API Server]
    end
    
    A --> G
    B --> G
    C -.-> G
    D -.-> G
    E -.-> G
    F -.-> G
```

---

## Docker Compose de Referência

```yaml
version: '3.8'

services:
  codeactions:
    build: .
    ports:
      - "8050:8050"
    environment:
      - FLOWS_CODE_ACTIONS_PORT=8050
      - FLOWS_CODE_ACTIONS_DB_TYPE=postgres
      - FLOWS_CODE_ACTIONS_DB_URI=postgres://postgres:postgres@db:5432/codeactions?sslmode=disable
      - FLOWS_CODE_ACTIONS_REDIS=redis://redis:6379/10
      - FLOWS_CODE_ACTIONS_S3_ENABLED=true
      - FLOWS_CODE_ACTIONS_S3_ENDPOINT=http://localstack:4566
      - FLOWS_CODE_ACTIONS_S3_BUCKET_NAME=codeactions
    depends_on:
      - db
      - redis
      - localstack

  db:
    image: postgres:15
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=codeactions
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      - SERVICES=s3
      - DEFAULT_REGION=us-east-1

volumes:
  postgres_data:
  redis_data:
```

---

## Validação de Configuração

O sistema valida configurações na inicialização:

```mermaid
flowchart TB
    A[Start] --> B[Load Config]
    B --> C{DB Type?}
    C -->|postgres| D[Connect PostgreSQL]
    C -->|mongodb| E[Connect MongoDB]
    D --> F{Success?}
    E --> F
    F -->|No| G[Fatal Error]
    F -->|Yes| H[Connect Redis]
    H --> I{Success?}
    I -->|No| G
    I -->|Yes| J{RabbitMQ URL?}
    J -->|Yes| K[Connect RabbitMQ]
    J -->|No| L[Skip EDA]
    K --> M{Success?}
    M -->|No| N[Log Warning]
    M -->|Yes| O[Start Consumers]
    L --> P[Setup Routes]
    N --> P
    O --> P
    P --> Q[Start Server]
```
