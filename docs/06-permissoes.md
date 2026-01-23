# 6. Sistema de Permissões

## Visão Geral

O sistema de permissões do Flows Code Actions controla o acesso aos recursos baseado em:
1. **Token de Autenticação**: Valida a identidade do usuário
2. **Permissões por Projeto**: Define o que o usuário pode fazer em cada projeto

## Fluxo de Autorização

```mermaid
flowchart TB
    A[Request] --> B{Auth Enabled?}
    B -->|No| C[Allow Request]
    B -->|Yes| D{Has Token?}
    D -->|No| E[401 Unauthorized]
    D -->|Yes| F{Valid Token?}
    F -->|No| E
    F -->|Yes| G[Extract User Email]
    G --> H{Check Permission}
    H -->|Has Permission| C
    H -->|No Permission| I[403 Forbidden]
```

## Modos de Autenticação

### 1. Auth Token Simples

Autenticação via token estático configurado em variável de ambiente.

```mermaid
sequenceDiagram
    participant Client
    participant Middleware
    participant Config
    
    Client->>Middleware: Request + Authorization: Bearer TOKEN
    Middleware->>Config: Get FLOWS_CODE_ACTIONS_AUTH_TOKEN
    Middleware->>Middleware: Compare Tokens
    
    alt Token Matches
        Middleware->>Middleware: Allow
    else Token Mismatch
        Middleware-->>Client: 401 Unauthorized
    end
```

**Configuração:**
```bash
FLOWS_CODE_ACTIONS_AUTH_TOKEN=seu-token-secreto
```

**Uso:**
```bash
curl -H "Authorization: Bearer seu-token-secreto" http://api/code
```

### 2. OIDC (OpenID Connect)

Autenticação via provedor OIDC (ex: Keycloak).

```mermaid
sequenceDiagram
    participant Client
    participant Middleware
    participant OIDC as OIDC Provider
    
    Client->>Middleware: Request + Authorization: Bearer JWT_TOKEN
    Middleware->>OIDC: Validate Token
    OIDC-->>Middleware: Token Claims
    Middleware->>Middleware: Extract email from claims
    
    alt Valid Token
        Middleware->>Middleware: Check Project Permission
    else Invalid Token
        Middleware-->>Client: 401 Unauthorized
    end
```

**Configuração:**
```bash
FLOWS_CODE_ACTIONS_OIDC_AUTH_ENABLED=true
FLOWS_CODE_ACTIONS_OIDC_REALM=seu-realm
FLOWS_CODE_ACTIONS_OIDC_HOST=https://keycloak.example.com
```

## Níveis de Permissão

### Roles (Papéis)

```mermaid
graph TB
    subgraph "Hierarquia de Roles"
        V[Viewer<br/>Role = 1]
        C[Contributor<br/>Role = 2]
        M[Moderator<br/>Role = 3]
    end
    
    V --> |Pode| R[Read]
    C --> |Pode| R
    M --> |Pode| R
    M --> |Pode| W[Write]
```

| Role | Valor | Descrição |
|------|-------|-----------|
| **Viewer** | 1 | Pode visualizar códigos e execuções |
| **Contributor** | 2 | Pode visualizar códigos e execuções |
| **Moderator** | 3 | Pode criar, editar e deletar códigos |

### Tipos de Permissão

| Permissão | Valor | Roles Permitidos |
|-----------|-------|------------------|
| **Read** | `read` | Viewer (1+), Contributor (2+), Moderator (3+) |
| **Write** | `write` | Moderator (3+) |

### Lógica de Verificação

```go
func HasPermission(user *UserPermission, permission PermissionAccess) bool {
    // Read: Qualquer role >= 1 tem acesso
    if permission == ReadPermission && user.Role >= 1 {
        return true
    }
    // Write: Apenas role >= 3 (Moderator) tem acesso
    if permission == WritePermission && user.Role >= 3 {
        return true
    }
    return false
}
```

## Endpoints e Permissões

### Matriz de Permissões por Endpoint

| Endpoint | Método | Permissão | Descrição |
|----------|--------|-----------|-----------|
| `/` | GET | - | Health check |
| `/health` | GET | - | Health check |
| `/healthz` | GET | - | Health check |
| `/code` | POST | Write | Criar código |
| `/code` | GET | Read | Listar códigos |
| `/code/:id` | GET | Read | Obter código |
| `/code/:id` | PATCH | Write | Atualizar código |
| `/code/:id` | DELETE | Write | Deletar código |
| `/admin/code` | POST | Write | Criar código (admin) |
| `/admin/code/:id` | PATCH | Write | Atualizar código (admin) |
| `/coderun/:id` | GET | Read | Obter execução |
| `/coderun` | GET | Read | Listar execuções |
| `/codelog/:id` | GET | Read | Obter log |
| `/codelog` | GET | Read | Listar logs |
| `/run/:code_id` | POST | Auth | Executar código |
| `/endpoint/:code_id` | ANY | - | Endpoint público |
| `/action/endpoint/:code_id` | ANY | - | Endpoint público (rate limited) |
| `/metrics` | GET | - | Métricas Prometheus |

## Middleware de Proteção

### ProtectEndpointWithAuthToken

Protege endpoints que requerem autenticação E verificação de permissão no projeto.

```go
func ProtectEndpointWithAuthToken(
    cfg *config.Config, 
    handler echo.HandlerFunc, 
    permission permission.PermissionAccess,
) echo.HandlerFunc
```

**Fluxo:**
```mermaid
flowchart TB
    A[Request] --> B{OIDC Enabled?}
    B -->|Yes| C[Validate OIDC Token]
    B -->|No| D{Auth Token Set?}
    D -->|Yes| E[Validate Auth Token]
    D -->|No| F[Skip Auth]
    
    C --> G[Extract User Email]
    E --> G
    F --> H[Execute Handler]
    
    G --> I[Check Permission in Project]
    I -->|OK| H
    I -->|Fail| J[403 Forbidden]
```

### RequireAuthToken

Protege endpoints que requerem apenas autenticação (sem verificação de permissão no projeto).

```go
func RequireAuthToken(cfg *config.Config, handler echo.HandlerFunc) echo.HandlerFunc
```

**Uso:** Endpoint `/run/:code_id`

## Gerenciamento de Permissões via RabbitMQ

O sistema recebe eventos de permissões via RabbitMQ:

```mermaid
flowchart LR
    subgraph "External System"
        A[Permission Service]
    end
    
    subgraph "RabbitMQ"
        B[Permission Exchange]
        C[Permission Queue]
    end
    
    subgraph "Code Actions"
        D[Permission Consumer]
        E[Permission Service]
        F[(Database)]
    end
    
    A -->|Publish Event| B
    B --> C
    C -->|Consume| D
    D -->|Create/Update/Delete| E
    E --> F
```

### Eventos Suportados

| Evento | Ação |
|--------|------|
| Permission Created | Cria nova permissão de usuário |
| Permission Updated | Atualiza role do usuário |
| Permission Deleted | Remove permissão do usuário |

## Configuração

### Variáveis de Ambiente

```bash
# Auth Token (modo simples)
FLOWS_CODE_ACTIONS_AUTH_TOKEN=token-secreto

# OIDC (modo avançado)
FLOWS_CODE_ACTIONS_OIDC_AUTH_ENABLED=true
FLOWS_CODE_ACTIONS_OIDC_REALM=meu-realm
FLOWS_CODE_ACTIONS_OIDC_HOST=https://keycloak.example.com

# RabbitMQ para eventos de permissão
FLOWS_CODE_ACTIONS_PERMISSION_EXCHANGE=permission-exchange
FLOW_CODE_ACTIONS_PERMISSION_QUEUE=permission-queue
```

## Verificação de Permissão no Código

### Exemplo de Uso

```go
// No handler
func (h *CodeHandler) CreateCode(c echo.Context) error {
    ctx := context.Background()
    projectUUID := c.QueryParam("project_uuid")
    
    // Verifica se o usuário tem permissão de escrita no projeto
    if err := CheckPermission(ctx, c, projectUUID); err != nil {
        return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
    }
    
    // Continua com a criação...
}
```

### Função CheckPermission

```go
func CheckPermission(ctx context.Context, c echo.Context, projectUUID string, permissionRole permission.PermissionAccess) error {
    // Se o sistema de permissões não está configurado, permite
    if Permission == nil {
        log.Info("auth permissions disabled")
        return nil
    }
    
    // Verifica a permissão
    err := Permission.CheckPermission(ctx, c, projectUUID, permissionRole)
    if err != nil {
        return echo.NewHTTPError(http.StatusForbidden, err)
    }
    return nil
}
```

## Diagrama de Decisão Completo

```mermaid
flowchart TB
    A[Request HTTP] --> B{Endpoint Público?}
    B -->|Sim| C[Processar Request]
    B -->|Não| D{Auth Configurada?}
    
    D -->|Não| C
    D -->|Sim| E{Modo OIDC?}
    
    E -->|Sim| F[Validar Token OIDC]
    E -->|Não| G[Validar Auth Token]
    
    F -->|Inválido| H[401 Unauthorized]
    G -->|Inválido| H
    
    F -->|Válido| I[Extrair Email]
    G -->|Válido| I
    
    I --> J{Endpoint requer<br/>permissão projeto?}
    J -->|Não| C
    J -->|Sim| K[Buscar UserPermission]
    
    K --> L{Encontrou?}
    L -->|Não| M[403 Forbidden]
    L -->|Sim| N{HasPermission?}
    
    N -->|Não| M
    N -->|Sim| C
    
    C --> O[Response]
```
