# 5. Migrações de Banco de Dados

## Visão Geral

O Flows Code Actions utiliza migrações SQL para gerenciar o schema do PostgreSQL. As migrações seguem o padrão do [golang-migrate](https://github.com/golang-migrate/migrate), com arquivos `up.sql` (aplicar migração) e `down.sql` (reverter migração).

### Estrutura de Arquivos

```
migrations/
├── 000001_create_codes_table.up.sql
├── 000001_create_codes_table.down.sql
├── 000002_create_codelibs_table.up.sql
├── 000002_create_codelibs_table.down.sql
├── 000003_create_coderuns_table.up.sql
├── 000003_create_coderuns_table.down.sql
├── 000004_create_user_permissions_table.up.sql
├── 000004_create_user_permissions_table.down.sql
├── 000005_create_projects_table.up.sql
├── 000005_create_projects_table.down.sql
└── README.md
```

### Ordem de Aplicação

```mermaid
flowchart LR
    A[000001<br/>codes] --> B[000002<br/>codelibs]
    B --> C[000003<br/>coderuns]
    C --> D[000004<br/>user_permissions]
    D --> E[000005<br/>projects]
```

---

## Migração 000001: Tabela `codes`

### Propósito
Cria a tabela principal que armazena os códigos executáveis do sistema.

### Diagrama da Tabela

```mermaid
erDiagram
    CODES {
        uuid id PK "gen_random_uuid()"
        text mongo_object_id UK "Compatibilidade MongoDB"
        varchar(255) name "NOT NULL"
        varchar(50) type "CHECK (flow, endpoint)"
        text source "NOT NULL"
        varchar(50) language "CHECK (python, go, javascript)"
        varchar(512) url "Opcional"
        varchar(255) project_uuid "NOT NULL"
        integer timeout "DEFAULT 60, CHECK 5-300"
        timestamp created_at "DEFAULT NOW()"
        timestamp updated_at "DEFAULT NOW()"
    }
```

### UP Migration

```sql
-- Habilita extensão para UUIDs
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Cria a tabela codes
CREATE TABLE IF NOT EXISTS codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('flow', 'endpoint')),
    source TEXT NOT NULL,
    language VARCHAR(50) NOT NULL CHECK (language IN ('python', 'go', 'javascript')),
    url VARCHAR(512),
    project_uuid VARCHAR(255) NOT NULL,
    timeout INTEGER NOT NULL DEFAULT 60 CHECK (timeout >= 5 AND timeout <= 300),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Índices Criados

| Índice | Colunas | Propósito |
|--------|---------|-----------|
| `idx_codes_mongo_object_id` | `mongo_object_id` | Busca por ID legado do MongoDB |
| `idx_codes_project_uuid` | `project_uuid` | Filtrar códigos por projeto |
| `idx_codes_type` | `type` | Filtrar por tipo (flow/endpoint) |
| `idx_codes_project_type` | `project_uuid, type` | Consultas combinadas projeto + tipo |
| `idx_codes_created_at` | `created_at` | Ordenação por data de criação |

### DOWN Migration

```sql
-- Remove índices
DROP INDEX IF EXISTS idx_codes_created_at;
DROP INDEX IF EXISTS idx_codes_project_type;
DROP INDEX IF EXISTS idx_codes_type;
DROP INDEX IF EXISTS idx_codes_project_uuid;
DROP INDEX IF EXISTS idx_codes_mongo_object_id;

-- Remove tabela
DROP TABLE IF EXISTS codes;
```

### Detalhes dos Campos

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `id` | UUID | Identificador único gerado automaticamente |
| `mongo_object_id` | TEXT | ObjectID do MongoDB para compatibilidade com migração de dados |
| `name` | VARCHAR(255) | Nome do código (obrigatório) |
| `type` | VARCHAR(50) | Tipo: `flow` (execução interna) ou `endpoint` (HTTP) |
| `source` | TEXT | Código fonte completo |
| `language` | VARCHAR(50) | Linguagem de programação |
| `url` | VARCHAR(512) | URL customizada (para endpoints) |
| `project_uuid` | VARCHAR(255) | UUID do projeto ao qual pertence |
| `timeout` | INTEGER | Timeout de execução em segundos (5-300) |
| `created_at` | TIMESTAMP | Data de criação |
| `updated_at` | TIMESTAMP | Data da última atualização |

---

## Migração 000002: Tabela `codelibs`

### Propósito
Armazena as bibliotecas externas disponíveis para uso nos códigos (atualmente apenas Python).

### Diagrama da Tabela

```mermaid
erDiagram
    CODELIBS {
        uuid id PK "gen_random_uuid()"
        text mongo_object_id UK "Compatibilidade MongoDB"
        varchar(255) name "NOT NULL"
        varchar(50) language "CHECK (python)"
        timestamp created_at "DEFAULT NOW()"
        timestamp updated_at "DEFAULT NOW()"
    }
```

### UP Migration

```sql
CREATE TABLE IF NOT EXISTS codelibs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    name VARCHAR(255) NOT NULL,
    language VARCHAR(50) NOT NULL CHECK (language IN ('python')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Índices Criados

| Índice | Colunas | Propósito |
|--------|---------|-----------|
| `idx_codelibs_mongo_object_id` | `mongo_object_id` | Busca por ID legado |
| `idx_codelibs_name` | `name` | Busca por nome da biblioteca |
| `idx_codelibs_language` | `language` | Filtrar por linguagem |
| `idx_codelibs_name_language` | `name, language` | Busca combinada |
| `idx_codelibs_created_at` | `created_at` | Ordenação temporal |
| `idx_codelibs_unique_name_language` | `name, language` | **UNIQUE** - Evita duplicatas |

### DOWN Migration

```sql
DROP INDEX IF EXISTS idx_codelibs_unique_name_language;
DROP INDEX IF EXISTS idx_codelibs_created_at;
DROP INDEX IF EXISTS idx_codelibs_name_language;
DROP INDEX IF EXISTS idx_codelibs_language;
DROP INDEX IF EXISTS idx_codelibs_name;
DROP INDEX IF EXISTS idx_codelibs_mongo_object_id;

DROP TABLE IF EXISTS codelibs;
```

### Observações
- A constraint `UNIQUE` em `(name, language)` garante que não existam bibliotecas duplicadas para a mesma linguagem
- Atualmente apenas `python` é suportado como linguagem

---

## Migração 000003: Tabela `coderuns`

### Propósito
Registra cada execução de código, incluindo parâmetros de entrada, resultado e metadados.

### Diagrama da Tabela

```mermaid
erDiagram
    CODERUNS {
        uuid id PK "gen_random_uuid()"
        text mongo_object_id UK "Compatibilidade MongoDB"
        uuid code_id "NOT NULL (ref codes)"
        text code_mongo_id "Ref MongoDB legado"
        varchar(50) status "CHECK (queued, started, completed, failed)"
        text result "Resultado da execução"
        jsonb extra "Metadados (status_code, content_type)"
        jsonb params "Parâmetros de entrada"
        text body "Body da requisição"
        jsonb headers "Headers HTTP"
        timestamp created_at "DEFAULT NOW()"
        timestamp updated_at "DEFAULT NOW()"
    }
```

### UP Migration

```sql
CREATE TABLE IF NOT EXISTS coderuns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    code_id UUID NOT NULL,
    code_mongo_id TEXT,
    status VARCHAR(50) NOT NULL CHECK (status IN ('queued', 'started', 'completed', 'failed')),
    result TEXT,
    extra JSONB DEFAULT '{}'::jsonb,
    params JSONB DEFAULT '{}'::jsonb,
    body TEXT,
    headers JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Índices Criados

| Índice | Tipo | Colunas | Propósito |
|--------|------|---------|-----------|
| `idx_coderuns_mongo_object_id` | B-tree | `mongo_object_id` | Busca por ID legado |
| `idx_coderuns_code_id` | B-tree | `code_id` | Buscar execuções de um código |
| `idx_coderuns_code_mongo_id` | B-tree | `code_mongo_id` | Compatibilidade MongoDB |
| `idx_coderuns_status` | B-tree | `status` | Filtrar por status |
| `idx_coderuns_created_at` | B-tree | `created_at` | Ordenação/limpeza |
| `idx_coderuns_code_id_created_at` | B-tree | `code_id, created_at` | Histórico por código |
| `idx_coderuns_extra` | **GIN** | `extra` | Consultas JSON |
| `idx_coderuns_params` | **GIN** | `params` | Consultas JSON |

### DOWN Migration

```sql
DROP INDEX IF EXISTS idx_coderuns_params;
DROP INDEX IF EXISTS idx_coderuns_extra;
DROP INDEX IF EXISTS idx_coderuns_code_id_created_at;
DROP INDEX IF EXISTS idx_coderuns_created_at;
DROP INDEX IF EXISTS idx_coderuns_status;
DROP INDEX IF EXISTS idx_coderuns_code_mongo_id;
DROP INDEX IF EXISTS idx_coderuns_code_id;
DROP INDEX IF EXISTS idx_coderuns_mongo_object_id;

DROP TABLE IF EXISTS coderuns;
```

### Detalhes dos Campos JSONB

#### Campo `extra`
```json
{
  "status_code": 200,      // Código HTTP de resposta
  "content_type": "json"   // Tipo de conteúdo (json, html, string)
}
```

#### Campo `params`
```json
{
  "param1": "valor1",
  "param2": 123
}
```

#### Campo `headers`
```json
{
  "Content-Type": ["application/json"],
  "Authorization": ["Bearer token123"]
}
```

### Observações
- Índices **GIN** são utilizados para campos JSONB, permitindo consultas eficientes dentro do JSON
- O campo `code_id` não possui FK para `codes` para permitir manter histórico mesmo após deleção do código

---

## Migração 000004: Tabela `user_permissions`

### Propósito
Define as permissões de usuários em projetos específicos.

### Diagrama da Tabela

```mermaid
erDiagram
    USER_PERMISSIONS {
        uuid id PK "gen_random_uuid()"
        text mongo_object_id UK "Compatibilidade MongoDB"
        varchar(255) project_uuid "NOT NULL"
        varchar(255) email "NOT NULL"
        integer role "NOT NULL (1=Viewer, 2=Contributor, 3=Moderator)"
        timestamp created_at "DEFAULT NOW()"
        timestamp updated_at "DEFAULT NOW()"
    }
```

### UP Migration

```sql
CREATE TABLE IF NOT EXISTS user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    project_uuid VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Índices Criados

| Índice | Colunas | Propósito |
|--------|---------|-----------|
| `idx_user_permissions_mongo_object_id` | `mongo_object_id` | Busca por ID legado |
| `idx_user_permissions_project_uuid` | `project_uuid` | Listar permissões de um projeto |
| `idx_user_permissions_email` | `email` | Buscar permissões de um usuário |
| `idx_user_permissions_project_email` | `project_uuid, email` | Busca específica |
| `idx_user_permissions_unique_project_email` | `project_uuid, email` | **UNIQUE** - Uma permissão por usuário/projeto |

### DOWN Migration

```sql
DROP INDEX IF EXISTS idx_user_permissions_unique_project_email;
DROP INDEX IF EXISTS idx_user_permissions_project_email;
DROP INDEX IF EXISTS idx_user_permissions_email;
DROP INDEX IF EXISTS idx_user_permissions_project_uuid;
DROP INDEX IF EXISTS idx_user_permissions_mongo_object_id;

DROP TABLE IF EXISTS user_permissions;
```

### Valores do Campo `role`

| Valor | Nome | Permissão Read | Permissão Write |
|-------|------|----------------|-----------------|
| 1 | Viewer | ✅ | ❌ |
| 2 | Contributor | ✅ | ❌ |
| 3 | Moderator | ✅ | ✅ |

---

## Migração 000005: Tabela `projects`

### Propósito
Armazena informações dos projetos que agrupam códigos e permissões.

### Diagrama da Tabela

```mermaid
erDiagram
    PROJECTS {
        uuid id PK "gen_random_uuid()"
        text mongo_object_id UK "Compatibilidade MongoDB"
        varchar(255) uuid UK "UUID do negócio"
        varchar(255) name "NOT NULL"
        jsonb authorizations "Array de autorizações"
        timestamp created_at "DEFAULT NOW()"
        timestamp updated_at "DEFAULT NOW()"
    }
```

### UP Migration

```sql
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mongo_object_id TEXT UNIQUE,
    uuid VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    authorizations JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Índices Criados

| Índice | Tipo | Colunas | Propósito |
|--------|------|---------|-----------|
| `idx_projects_mongo_object_id` | B-tree | `mongo_object_id` | Busca por ID legado |
| `idx_projects_uuid` | B-tree | `uuid` | Busca por UUID do projeto |
| `idx_projects_name` | B-tree | `name` | Busca por nome |
| `idx_projects_created_at` | B-tree | `created_at` | Ordenação temporal |
| `idx_projects_authorizations` | **GIN** | `authorizations` | Consultas JSON |

### DOWN Migration

```sql
DROP INDEX IF EXISTS idx_projects_authorizations;
DROP INDEX IF EXISTS idx_projects_created_at;
DROP INDEX IF EXISTS idx_projects_name;
DROP INDEX IF EXISTS idx_projects_uuid;
DROP INDEX IF EXISTS idx_projects_mongo_object_id;

DROP TABLE IF EXISTS projects;
```

### Estrutura do Campo `authorizations`

```json
[
  {
    "user_email": "admin@empresa.com",
    "role": "admin"
  },
  {
    "user_email": "dev@empresa.com",
    "role": "contributor"
  }
]
```

---

## Diagrama Completo do Schema

```mermaid
erDiagram
    PROJECTS ||--o{ CODES : "contém"
    PROJECTS ||--o{ USER_PERMISSIONS : "define"
    CODES ||--o{ CODERUNS : "gera"
    
    PROJECTS {
        uuid id PK
        text mongo_object_id UK
        varchar uuid UK
        varchar name
        jsonb authorizations
        timestamp created_at
        timestamp updated_at
    }
    
    CODES {
        uuid id PK
        text mongo_object_id UK
        varchar name
        varchar type
        text source
        varchar language
        varchar url
        varchar project_uuid FK
        integer timeout
        timestamp created_at
        timestamp updated_at
    }
    
    CODERUNS {
        uuid id PK
        text mongo_object_id UK
        uuid code_id FK
        text code_mongo_id
        varchar status
        text result
        jsonb extra
        jsonb params
        text body
        jsonb headers
        timestamp created_at
        timestamp updated_at
    }
    
    USER_PERMISSIONS {
        uuid id PK
        text mongo_object_id UK
        varchar project_uuid FK
        varchar email
        integer role
        timestamp created_at
        timestamp updated_at
    }
    
    CODELIBS {
        uuid id PK
        text mongo_object_id UK
        varchar name
        varchar language
        timestamp created_at
        timestamp updated_at
    }
```

---

## Comandos de Migração

### Aplicar todas as migrações
```bash
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/codeactions?sslmode=disable" up
```

### Reverter última migração
```bash
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/codeactions?sslmode=disable" down 1
```

### Reverter todas as migrações
```bash
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/codeactions?sslmode=disable" down
```

### Verificar versão atual
```bash
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/codeactions?sslmode=disable" version
```

### Forçar versão (em caso de erro)
```bash
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/codeactions?sslmode=disable" force VERSION
```
