# 9. Diagramas e Flowcharts

## Arquitetura Completa do Sistema

```mermaid
flowchart TB
    subgraph "Clientes"
        U1[Usuários]
        U2[Chatbots]
        U3[Sistemas Externos]
    end
    
    subgraph "Load Balancer"
        LB[Nginx/ALB]
    end
    
    subgraph "Flows Code Actions"
        subgraph "HTTP Layer"
            E[Echo Server]
            MW[Middlewares]
            RL[Rate Limiter]
        end
        
        subgraph "Handlers"
            CH[Code Handler]
            CRH[CodeRun Handler]
            CLH[CodeLog Handler]
            CRUNH[CodeRunner Handler]
            HH[Health Handler]
        end
        
        subgraph "Services"
            CS[Code Service]
            CRS[CodeRun Service]
            CLS[CodeLog Service]
            CRUNS[CodeRunner Service]
            PS[Permission Service]
            PRS[Project Service]
        end
        
        subgraph "Repositories"
            direction LR
            R1[Code Repo]
            R2[CodeRun Repo]
            R3[CodeLog Repo]
            R4[Permission Repo]
            R5[Project Repo]
        end
        
        subgraph "Code Execution"
            PY[Python Engine]
            JS[JavaScript Engine]
            GO[Go Engine]
            CG[CGroups Manager]
        end
        
        subgraph "EDA"
            PC[Project Consumer]
            PMC[Permission Consumer]
        end
    end
    
    subgraph "Data Stores"
        PG[(PostgreSQL)]
        MG[(MongoDB)]
        RD[(Redis)]
        S3[(S3/LocalStack)]
    end
    
    subgraph "External Services"
        RMQ[RabbitMQ]
        OIDC[OIDC Provider]
        SEN[Sentry]
    end
    
    U1 --> LB
    U2 --> LB
    U3 --> LB
    LB --> E
    
    E --> MW
    MW --> RL
    RL --> CH
    RL --> CRH
    RL --> CLH
    RL --> CRUNH
    RL --> HH
    
    CH --> CS
    CRH --> CRS
    CLH --> CLS
    CRUNH --> CRUNS
    
    CS --> R1
    CRS --> R2
    CLS --> R3
    PS --> R4
    PRS --> R5
    
    R1 --> PG
    R1 --> MG
    R2 --> PG
    R2 --> MG
    R3 --> S3
    R3 --> MG
    R4 --> PG
    R4 --> MG
    R5 --> PG
    R5 --> MG
    
    CRUNS --> PY
    CRUNS --> JS
    CRUNS --> GO
    CRUNS --> CG
    
    CRUNS --> CRS
    CRUNS --> CLS
    
    RMQ --> PC
    RMQ --> PMC
    PC --> PRS
    PC --> PS
    PMC --> PS
    
    MW --> RD
    RL --> RD
    
    MW -.-> OIDC
    E -.-> SEN
```

---

## Fluxo de Execução de Código Python

```mermaid
flowchart TB
    A[Request POST /action/endpoint/:code_id] --> B{Rate Limit OK?}
    B -->|No| C[429 Too Many Requests]
    B -->|Yes| D[Get Code from DB]
    D --> E{Code Exists?}
    E -->|No| F[404 Not Found]
    E -->|Yes| G{Type = endpoint?}
    G -->|No| F
    G -->|Yes| H[Create CodeRun - status: started]
    
    H --> I[Create Temp Directory]
    I --> J[Copy main.py Engine]
    J --> K[Write User Code as action.py]
    K --> L[Prepare Arguments]
    L --> M[Execute Python Process]
    
    M --> N{CGroups Enabled?}
    N -->|Yes| O[Add Process to CGroup]
    N -->|No| P[Continue]
    O --> P
    
    P --> Q[Wait for Process]
    Q --> R{Timeout?}
    R -->|Yes| S[Kill Process]
    R -->|No| T{Error in stderr?}
    
    S --> U[Update CodeRun - status: failed]
    T -->|Yes| U
    T -->|No| V[Parse Result from stdout]
    
    V --> W[Update CodeRun - status: completed]
    W --> X[Extract status_code from result]
    X --> Y[Extract content_type from result]
    Y --> Z[Set Response Headers]
    Z --> AA[Return Response]
    
    U --> AB[Return Error Response]
    
    subgraph "Cleanup"
        AC[Remove Temp Directory]
    end
    
    AA --> AC
    AB --> AC
```

---

## Fluxo de Autenticação e Autorização

```mermaid
flowchart TB
    A[HTTP Request] --> B{Endpoint Protected?}
    B -->|No| C[Process Request]
    B -->|Yes| D{Auth Mode?}
    
    D -->|Token| E[Extract Bearer Token]
    D -->|OIDC| F[Extract JWT Token]
    
    E --> G{Token Valid?}
    F --> H[Validate with OIDC Provider]
    
    G -->|No| I[401 Unauthorized]
    H -->|Invalid| I
    
    G -->|Yes| J[Extract User Info]
    H -->|Valid| J
    
    J --> K{Requires Project Permission?}
    K -->|No| C
    K -->|Yes| L[Get project_uuid from Request]
    
    L --> M[Query UserPermission]
    M --> N{Permission Found?}
    N -->|No| O[403 Forbidden]
    N -->|Yes| P{Has Required Role?}
    
    P -->|No| O
    P -->|Yes| C
    
    C --> Q[Execute Handler]
    Q --> R[Return Response]
```

---

## Ciclo de Vida do CodeRun

```mermaid
stateDiagram-v2
    [*] --> Created: Request received
    
    Created --> Queued: Save to DB
    Queued --> Started: Begin execution
    
    Started --> Completed: Success
    Started --> Failed: Error/Timeout
    
    Completed --> [*]: Response sent
    Failed --> [*]: Error response
    
    note right of Started
        Logs são salvos
        durante execução
    end note
    
    note right of Completed
        Result e Extra
        são atualizados
    end note
    
    note right of Failed
        Error message
        salvo no result
    end note
```

---

## Fluxo de Eventos RabbitMQ

```mermaid
sequenceDiagram
    participant Ext as Sistema Externo
    participant RMQ as RabbitMQ
    participant CON as Consumer
    participant SVC as Service
    participant DB as Database
    
    rect rgb(200, 220, 240)
        Note over Ext,DB: Criação de Projeto
        Ext->>RMQ: Publish {uuid, name, authorizations}
        RMQ->>CON: Deliver message
        CON->>SVC: FindByUUID(uuid)
        SVC->>DB: SELECT
        DB-->>SVC: null
        CON->>SVC: Create(project)
        SVC->>DB: INSERT
        
        loop Para cada authorization
            CON->>SVC: Create(permission)
            SVC->>DB: INSERT permission
        end
        
        CON->>RMQ: ACK
    end
    
    rect rgb(220, 240, 200)
        Note over Ext,DB: Atualização de Permissão
        Ext->>RMQ: Publish {action: update, email, role}
        RMQ->>CON: Deliver message
        CON->>SVC: Find(project_uuid, email)
        SVC->>DB: SELECT
        DB-->>SVC: permission
        CON->>SVC: Update(permission)
        SVC->>DB: UPDATE
        CON->>RMQ: ACK
    end
    
    rect rgb(240, 220, 200)
        Note over Ext,DB: Remoção de Permissão
        Ext->>RMQ: Publish {action: delete, email}
        RMQ->>CON: Deliver message
        CON->>SVC: Find(project_uuid, email)
        SVC->>DB: SELECT
        DB-->>SVC: permission
        CON->>SVC: Delete(id)
        SVC->>DB: DELETE
        CON->>RMQ: ACK
    end
```

---

## Estrutura do Python Engine

```mermaid
flowchart LR
    subgraph "Temp Directory"
        A[main.py<br/>Engine principal]
        B[action.py<br/>Código do usuário]
    end
    
    subgraph "Execution"
        C[python main.py]
    end
    
    subgraph "Arguments"
        D[-a params JSON]
        E[-b body]
        F[-r run_id]
        G[-c code_id]
        H[-H headers JSON]
    end
    
    subgraph "Output"
        I[stdout: result]
        J[stderr: errors]
    end
    
    A --> C
    B --> C
    D --> C
    E --> C
    F --> C
    G --> C
    H --> C
    C --> I
    C --> J
```

---

## Estrutura de Dados - Relacionamentos

```mermaid
erDiagram
    PROJECT ||--o{ CODE : "1:N"
    PROJECT ||--o{ USER_PERMISSION : "1:N"
    CODE ||--o{ CODE_RUN : "1:N"
    CODE_RUN ||--o{ CODE_LOG : "1:N"
    CODE_LIB }o--o{ CODE : "N:M (implicit)"
    
    PROJECT {
        uuid id PK
        varchar uuid UK
        varchar name
        jsonb authorizations
    }
    
    CODE {
        uuid id PK
        varchar name
        enum type "flow|endpoint"
        text source
        enum language "python|javascript|go"
        varchar project_uuid FK
        int timeout
    }
    
    CODE_RUN {
        uuid id PK
        uuid code_id FK
        enum status "queued|started|completed|failed"
        text result
        jsonb extra
        jsonb params
        text body
        jsonb headers
    }
    
    CODE_LOG {
        string id PK
        string run_id FK
        string code_id FK
        enum type "debug|info|error"
        text content
    }
    
    USER_PERMISSION {
        uuid id PK
        varchar project_uuid FK
        varchar email
        int role "1=Viewer|2=Contributor|3=Moderator"
    }
    
    CODE_LIB {
        uuid id PK
        varchar name
        enum language "python"
    }
```

---

## Rate Limiting Flow

```mermaid
flowchart TB
    A[Request /action/endpoint/:code_id] --> B[Extract code_id]
    B --> C[Generate Redis Key]
    C --> D{Key Exists?}
    
    D -->|No| E[Create Key with count=1]
    D -->|Yes| F[Increment Count]
    
    E --> G[Set TTL = window_seconds]
    F --> H{Count > Max?}
    
    G --> I[Allow Request]
    H -->|No| I
    H -->|Yes| J[429 Too Many Requests]
    
    I --> K[Process Request]
    
    subgraph "Redis"
        L["key: rate_limit:{code_id}"]
        M["value: request_count"]
        N["ttl: 60 seconds"]
    end
```

---

## Cleaner Tasks Flow

```mermaid
flowchart TB
    subgraph "Server Startup"
        A[Start Server] --> B[Initialize Services]
        B --> C[Start CodeLog Cleaner]
        B --> D[Start CodeRun Cleaner]
    end
    
    subgraph "CodeLog Cleaner"
        C --> E{Obtain Redis Lock?}
        E -->|No| F[Skip - Another Instance Running]
        E -->|Yes| G[Schedule Daily Task]
        G --> H[At scheduled time]
        H --> I[Delete logs older than retention_period]
        I --> J[Release Lock after 1 hour]
        J --> G
    end
    
    subgraph "CodeRun Cleaner"
        D --> K{Obtain Redis Lock?}
        K -->|No| L[Skip - Another Instance Running]
        K -->|Yes| M[Schedule Daily Task]
        M --> N[At scheduled time]
        N --> O[Delete runs older than retention_period]
        O --> P[Release Lock after 1 hour]
        P --> M
    end
```

---

## Metrics Collection

```mermaid
flowchart LR
    subgraph "Code Actions"
        A[Handler] --> B[Metrics Middleware]
        B --> C[Prometheus Metrics]
    end
    
    subgraph "Metrics"
        C --> D[code_created_total]
        C --> E[code_run_total]
        C --> F[code_run_duration_seconds]
    end
    
    subgraph "Prometheus"
        G[GET /metrics] --> H[Scrape Metrics]
    end
    
    C --> G
    
    subgraph "Grafana"
        H --> I[Dashboards]
        H --> J[Alerts]
    end
```

---

## Deployment Architecture

```mermaid
flowchart TB
    subgraph "Kubernetes Cluster"
        subgraph "Namespace: codeactions"
            subgraph "Deployments"
                D1[codeactions-api<br/>replicas: 3]
            end
            
            subgraph "Services"
                S1[codeactions-svc<br/>ClusterIP]
            end
            
            subgraph "Ingress"
                I1[codeactions-ingress<br/>api.example.com]
            end
            
            subgraph "ConfigMaps"
                CM1[codeactions-config]
            end
            
            subgraph "Secrets"
                SEC1[codeactions-secrets]
            end
        end
        
        subgraph "Namespace: databases"
            PG[(PostgreSQL)]
            RD[(Redis)]
        end
        
        subgraph "Namespace: messaging"
            RMQ[RabbitMQ]
        end
    end
    
    subgraph "External"
        ALB[AWS ALB]
        S3[(S3 Bucket)]
        OIDC[Keycloak]
    end
    
    ALB --> I1
    I1 --> S1
    S1 --> D1
    D1 --> PG
    D1 --> RD
    D1 --> RMQ
    D1 --> S3
    D1 -.-> OIDC
    
    CM1 --> D1
    SEC1 --> D1
```
