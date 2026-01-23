# 1. Visão Geral

## O que é o Flows Code Actions?

O **Flows Code Actions** é um serviço de execução de código dinâmico que permite aos usuários criar, gerenciar e executar código personalizado em Python, JavaScript ou Go. O sistema é projetado para ser integrado a fluxos de automação (como chatbots) ou para expor funcionalidades customizadas como endpoints HTTP.

## Casos de Uso

### 1. Automação de Fluxos (Type: `flow`)
Código executado internamente como parte de um fluxo de automação. Ideal para:
- Processamento de dados
- Integrações com APIs externas
- Transformações de dados
- Lógica de negócio customizada

### 2. Endpoints HTTP (Type: `endpoint`)
Código exposto como endpoint HTTP público. Ideal para:
- Webhooks
- APIs customizadas
- Integrações com sistemas externos
- Microserviços leves

## Fluxo de Funcionamento Geral

```mermaid
flowchart TB
    subgraph Cliente
        A[Usuário/Sistema]
    end
    
    subgraph "Flows Code Actions API"
        B[HTTP Server<br/>Echo Framework]
        C[Handlers]
        D[Services]
        E[Repositories]
    end
    
    subgraph "Armazenamento"
        F[(PostgreSQL)]
        G[(MongoDB<br/>Legado)]
        H[(Redis)]
        I[(S3)]
    end
    
    subgraph "Execução"
        J[Code Runner]
        K[Python Engine]
        L[JavaScript Engine]
        M[Go Engine]
    end
    
    subgraph "Mensageria"
        N[RabbitMQ]
        O[Project Consumer]
        P[Permission Consumer]
    end
    
    A -->|HTTP Request| B
    B --> C
    C --> D
    D --> E
    E --> F
    E --> G
    D --> H
    D --> J
    J --> K
    J --> L
    J --> M
    K -->|Logs| I
    N --> O
    N --> P
    O --> D
    P --> D
```

## Ciclo de Vida de um Code Action

### Criação

```mermaid
sequenceDiagram
    participant User as Usuário
    participant API as API
    participant Auth as Auth Handler
    participant Service as Code Service
    participant DB as Database
    
    User->>API: POST /code
    API->>Auth: Verificar Token
    Auth->>Auth: Validar Permissões
    Auth-->>API: Autorizado
    API->>Service: Create(code)
    Service->>Service: Validar Type e Language
    Service->>DB: Insert Code
    DB-->>Service: Code criado
    Service-->>API: Code Response
    API-->>User: 201 Created
```

### Execução

```mermaid
sequenceDiagram
    participant User as Usuário
    participant API as API
    participant Runner as Code Runner
    participant DB as Database
    participant S3 as S3 Storage
    
    User->>API: POST /run/{code_id}
    API->>DB: Get Code by ID
    DB-->>API: Code
    API->>Runner: RunCode(code)
    Runner->>DB: Create CodeRun (status: started)
    Runner->>Runner: Execute Code
    Runner->>S3: Save Logs
    Runner->>DB: Update CodeRun (status: completed)
    Runner-->>API: Result
    API-->>User: 200 OK + Result
```

## Principais Componentes

### 1. Code (Código)
Representa um código que pode ser executado. Contém:
- Nome e identificador
- Código fonte
- Linguagem (python, javascript, go)
- Tipo (flow, endpoint)
- Projeto associado
- Timeout de execução

### 2. CodeRun (Execução)
Representa uma execução de um código. Contém:
- Referência ao código
- Status (queued, started, completed, failed)
- Resultado da execução
- Parâmetros e headers utilizados
- Timestamps

### 3. CodeLog (Log)
Logs gerados durante a execução de um código. Contém:
- Referência ao CodeRun
- Tipo de log (debug, info, error)
- Conteúdo do log

### 4. CodeLib (Biblioteca)
Bibliotecas disponíveis para uso no código. Atualmente suporta:
- Bibliotecas Python (pip packages)

### 5. Project (Projeto)
Agrupa códigos e gerencia autorizações. Contém:
- UUID único
- Nome
- Lista de autorizações

### 6. UserPermission (Permissão)
Define permissões de usuários em projetos. Contém:
- Projeto associado
- Email do usuário
- Papel (Viewer=1, Contributor=2, Moderator=3)

## Linguagens Suportadas

| Linguagem | Status | Observações |
|-----------|--------|-------------|
| **Python** | ✅ Completo | Suporte a bibliotecas externas via pip |
| **JavaScript** | ✅ Básico | Execução via Node.js |
| **Go** | ✅ Básico | Compilação e execução dinâmica |

## Limites e Configurações

| Parâmetro | Valor Padrão | Descrição |
|-----------|--------------|-----------|
| Timeout mínimo | 5 segundos | Tempo mínimo de execução |
| Timeout máximo | 300 segundos | Tempo máximo de execução |
| Timeout padrão | 60 segundos | Timeout se não especificado |
| Rate Limit | 600 req/60s | Limite por code_id |
| Retenção de Logs | 30 dias | Período de retenção padrão |
