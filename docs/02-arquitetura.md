# 2. Arquitetura

## Arquitetura em Camadas

O Flows Code Actions segue uma arquitetura em camadas bem definida:

```mermaid
flowchart TB
    subgraph "Camada de Apresentação"
        A[HTTP Handlers]
        B[Middlewares]
        C[Rate Limiter]
    end
    
    subgraph "Camada de Aplicação"
        D[Use Cases / Services]
        E[Validators]
    end
    
    subgraph "Camada de Domínio"
        F[Entities]
        G[Interfaces]
    end
    
    subgraph "Camada de Infraestrutura"
        H[PostgreSQL Repository]
        I[MongoDB Repository]
        J[S3 Repository]
        K[Redis Client]
        L[RabbitMQ Client]
    end
    
    A --> D
    B --> A
    C --> A
    D --> F
    D --> G
    G --> H
    G --> I
    G --> J
    D --> K
    D --> L
```

## Componentes Principais

### 1. HTTP Server (Echo)

O servidor HTTP utiliza o framework Echo e implementa:

```mermaid
flowchart LR
    subgraph "Echo Server"
        A[Request] --> B[Logger Middleware]
        B --> C[CORS Middleware]
        C --> D[Auth Middleware]
        D --> E[Rate Limiter]
        E --> F[Handler]
        F --> G[Response]
    end
```

**Middlewares:**
- **Request Logger**: Log de todas as requisições
- **CORS**: Configuração de Cross-Origin
- **Auth Token**: Validação de tokens de autenticação
- **Rate Limiter**: Limite de requisições por código

### 2. Padrão Repository

Cada entidade segue o padrão Repository para abstração do banco de dados:

```mermaid
classDiagram
    class Repository {
        <<interface>>
        +Create(ctx, entity) entity
        +GetByID(ctx, id) entity
        +Update(ctx, id, entity) entity
        +Delete(ctx, id) error
    }
    
    class PostgresRepository {
        -db *sql.DB
        +Create(ctx, entity) entity
        +GetByID(ctx, id) entity
    }
    
    class MongoRepository {
        -db *mongo.Database
        +Create(ctx, entity) entity
        +GetByID(ctx, id) entity
    }
    
    Repository <|.. PostgresRepository
    Repository <|.. MongoRepository
```

### 3. Code Runner

O Code Runner é responsável por executar código em diferentes linguagens:

```mermaid
flowchart TB
    A[RunCode Request] --> B{Language?}
    
    B -->|Python| C[Python Runner]
    B -->|JavaScript| D[Node.js Runner]
    B -->|Go| E[Go Runner]
    
    C --> F[Create Temp Dir]
    F --> G[Copy Engine Files]
    G --> H[Write User Code]
    H --> I[Execute Process]
    I --> J{CGroup Enabled?}
    J -->|Yes| K[Apply Resource Limits]
    J -->|No| L[Continue]
    K --> L
    L --> M[Capture Output]
    M --> N[Cleanup]
    N --> O[Return Result]
    
    D --> P[Execute node -e code]
    P --> Q[Capture Output]
    Q --> O
    
    E --> R[Create Temp Dir]
    R --> S[Write Go Code]
    S --> T[go run]
    T --> U[Capture Output]
    U --> O
```

### 4. Event-Driven Architecture

O sistema utiliza RabbitMQ para eventos assíncronos:

```mermaid
flowchart LR
    subgraph "Producers"
        A[External Services]
    end
    
    subgraph "RabbitMQ"
        B[Project Exchange]
        C[Permission Exchange]
        D[Project Queue]
        E[Permission Queue]
    end
    
    subgraph "Consumers"
        F[Project Consumer]
        G[Permission Consumer]
    end
    
    subgraph "Services"
        H[Project Service]
        I[Permission Service]
    end
    
    A --> B
    A --> C
    B --> D
    C --> E
    D --> F
    E --> G
    F --> H
    G --> I
```

## Fluxo de Dados

### Criação de Código

```mermaid
sequenceDiagram
    participant Client
    participant Handler as Code Handler
    participant Auth as Auth Middleware
    participant Service as Code Service
    participant Repo as Repository
    participant DB as Database
    
    Client->>Handler: POST /code
    Handler->>Auth: Validate Token
    Auth-->>Handler: OK
    Handler->>Handler: Parse Query Params
    Handler->>Handler: Read Body (source)
    Handler->>Service: Create(code)
    Service->>Service: Validate Type
    Service->>Service: Validate Language
    Service->>Service: Check Blacklist
    Service->>Repo: Create(code)
    Repo->>DB: INSERT
    DB-->>Repo: New Code
    Repo-->>Service: Code
    Service-->>Handler: Code
    Handler-->>Client: 201 Created
```

### Execução de Código (Action Endpoint)

```mermaid
sequenceDiagram
    participant Client
    participant Handler as Runner Handler
    participant RateLimiter as Rate Limiter
    participant Service as Code Service
    participant Runner as Code Runner
    participant RunRepo as CodeRun Repository
    participant LogRepo as CodeLog Repository
    
    Client->>Handler: POST /action/endpoint/{code_id}
    Handler->>RateLimiter: Check Limit
    RateLimiter-->>Handler: OK
    Handler->>Service: GetByID(code_id)
    Service-->>Handler: Code
    Handler->>Runner: RunCode(code, params, body, headers)
    Runner->>RunRepo: Create CodeRun (started)
    Runner->>Runner: Execute Python/JS/Go
    Runner->>LogRepo: Save Logs
    Runner->>RunRepo: Update CodeRun (completed)
    Runner-->>Handler: Result
    Handler->>Handler: Set Content-Type
    Handler-->>Client: Response
```

## Gerenciamento de Recursos

O sistema suporta CGroups para limitar recursos de execução:

```mermaid
flowchart TB
    subgraph "Resource Management"
        A[Config] --> B{Enabled?}
        B -->|Yes| C[InitCGroup]
        C --> D[Load or Create CGroup]
        D --> E[Apply CPU Limits]
        D --> F[Apply Memory Limits]
        E --> G[Add Process to CGroup]
        F --> G
        B -->|No| H[Run Without Limits]
    end
```

**Configurações disponíveis:**
- `CPU Shares`: Proporção de CPU
- `CPU Quota`: Limite absoluto de CPU
- `Memory Limit`: Limite de memória
- `Memory Reservation`: Reserva de memória

## Armazenamento

### PostgreSQL (Principal)
- Codes
- CodeRuns
- CodeLibs
- UserPermissions
- Projects

### MongoDB (Legado)
- Suporte mantido para compatibilidade
- Mesmas entidades do PostgreSQL

### S3 (Logs)
- CodeLogs armazenados como objetos
- Organização por prefixo configurável
- Suporte a LocalStack para desenvolvimento

### Redis
- Cache de sessões
- Rate limiting
- Locks distribuídos para tarefas de limpeza

## Cleaners (Tarefas de Limpeza)

O sistema executa tarefas periódicas de limpeza:

```mermaid
flowchart TB
    A[Server Start] --> B[StartCodeLogCleaner]
    A --> C[StartCodeRunCleaner]
    
    B --> D{Obtain Lock?}
    C --> E{Obtain Lock?}
    
    D -->|Yes| F[Delete Old Logs]
    D -->|No| G[Skip - Other Instance Running]
    
    E -->|Yes| H[Delete Old Runs]
    E -->|No| I[Skip - Other Instance Running]
    
    F --> J[Schedule Next Run]
    H --> K[Schedule Next Run]
```

**Configurações:**
- `ScheduleTime`: Horário de execução (padrão: 01:00)
- `RetentionPeriod`: Período de retenção em dias (padrão: 30)
