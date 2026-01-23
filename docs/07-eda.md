# 7. Event-Driven Architecture (EDA)

## Visão Geral

O Flows Code Actions utiliza RabbitMQ para receber eventos de sistemas externos, permitindo sincronização automática de projetos e permissões.

## Arquitetura

```mermaid
flowchart TB
    subgraph "Sistemas Externos"
        A[Project Service]
        B[Permission Service]
    end
    
    subgraph "RabbitMQ"
        C[Project Exchange]
        D[Permission Exchange]
        E[Project Queue]
        F[Permission Queue]
    end
    
    subgraph "Flows Code Actions"
        G[EDA Manager]
        H[Project Consumer]
        I[Permission Consumer]
        J[Project Service]
        K[Permission Service]
        L[(PostgreSQL)]
        M[(MongoDB)]
    end
    
    A -->|Publish| C
    B -->|Publish| D
    C --> E
    D --> F
    E -->|Consume| H
    F -->|Consume| I
    G -.->|Manage| H
    G -.->|Manage| I
    H --> J
    I --> K
    J --> L
    J --> M
    K --> L
    K --> M
```

## Componentes

### EDA Manager

O `EDA` é responsável por gerenciar os consumers de eventos:

```go
type EDA struct {
    URL       string
    Consumers []*rabbitmq.Consumer
}

// Adiciona um consumer
func (e *EDA) AddConsumer(consumer *rabbitmq.Consumer)

// Inicia todos os consumers
func (e *EDA) StartConsumers() error
```

### Consumer

Interface genérica para consumo de mensagens:

```go
type Consumer struct {
    ExchangeName string           // Nome do exchange
    QueueName    string           // Nome da queue
    Handler      IConsumerHandler // Handler de processamento
}

type IConsumerHandler interface {
    Handle(context.Context, []byte) error
}
```

## Consumers Implementados

### 1. Project Consumer

Processa eventos relacionados a projetos.

```mermaid
sequenceDiagram
    participant RabbitMQ
    participant Consumer as Project Consumer
    participant Handler as Project Handler
    participant Service as Project Service
    participant PermService as Permission Service
    participant DB as Database
    
    RabbitMQ->>Consumer: Message
    Consumer->>Handler: Handle(msg)
    Handler->>Handler: Parse JSON
    Handler->>Handler: Detect Action
    
    alt Create Project
        Handler->>Service: Create(project)
        Service->>DB: Insert
        Handler->>PermService: Create Permissions
    else Update Project
        Handler->>Service: Update(project)
        Service->>DB: Update
    end
    
    Handler-->>Consumer: Success
    Consumer->>RabbitMQ: ACK
```

**Estrutura da Mensagem:**

```json
{
  "uuid": "proj-123-456",
  "name": "Meu Projeto",
  "authorizations": [
    {
      "user_email": "admin@empresa.com",
      "role": "admin"
    }
  ]
}
```

**Ações:**
- Se o projeto não existe: Cria projeto + permissões iniciais
- Se o projeto existe: Atualiza dados do projeto

### 2. Permission Consumer

Processa eventos de permissões de usuários.

```mermaid
sequenceDiagram
    participant RabbitMQ
    participant Consumer as Permission Consumer
    participant Handler as Permission Handler
    participant Service as Permission Service
    participant DB as Database
    
    RabbitMQ->>Consumer: Message
    Consumer->>Handler: Handle(msg)
    Handler->>Handler: Parse JSON
    Handler->>Handler: Detect Action
    
    alt Create Permission
        Handler->>Service: Create(permission)
        Service->>DB: Insert
    else Update Permission
        Handler->>Service: Update(permission)
        Service->>DB: Update
    else Delete Permission
        Handler->>Service: Delete(id)
        Service->>DB: Delete
    end
    
    Handler-->>Consumer: Success
    Consumer->>RabbitMQ: ACK
```

**Estrutura da Mensagem:**

```json
{
  "action": "create",
  "project_uuid": "proj-123-456",
  "email": "usuario@empresa.com",
  "role": 3
}
```

**Actions Suportadas:**
| Action | Descrição |
|--------|-----------|
| `create` | Cria nova permissão |
| `update` | Atualiza role existente |
| `delete` | Remove permissão |

## Configuração do RabbitMQ

### Variáveis de Ambiente

```bash
# URL de conexão
FLOWS_CODE_ACTIONS_RABBITMQ_URL=amqp://user:pass@localhost:5672

# Project events
FLOWS_CODE_ACTIONS_PROJECT_EXCHANGE=projects
FLOWS_CODE_ACTIONS_PROJECT_QUEUE=codeactions-projects

# Permission events
FLOWS_CODE_ACTIONS_PERMISSION_EXCHANGE=permissions
FLOW_CODE_ACTIONS_PERMISSION_QUEUE=codeactions-permissions
```

### Declaração de Exchanges e Queues

O sistema utiliza `ExchangeDeclarePassive` e `QueueDeclarePassive`, ou seja, espera que exchanges e queues já existam no RabbitMQ.

```go
// Exchange deve ser do tipo "topic"
err := ch.ExchangeDeclarePassive(
    c.ExchangeName,
    "topic",    // tipo
    true,       // durable
    false,      // auto-delete
    false,      // internal
    false,      // no-wait
    nil,
)

// Queue binding com "#" (recebe todas as mensagens)
err = ch.QueueBind(
    c.QueueName,
    "#",                // routing key (todas as mensagens)
    c.ExchangeName,
    false,
    nil,
)
```

## Tratamento de Erros

```mermaid
flowchart TB
    A[Mensagem Recebida] --> B[Handle Message]
    B --> C{Erro?}
    C -->|Não| D[ACK]
    C -->|Sim| E{Tipo de Erro?}
    E -->|ErrInvalidMsg| F[ACK - Descarta Mensagem]
    E -->|Outro Erro| G[Return Error]
    G --> H[Consumer Reconnect]
```

**Tipos de Erro:**
- `ErrInvalidMsg`: Mensagem inválida - ACK é enviado para descartar
- Outros erros: Causa reconexão do consumer

## Fluxo de Inicialização

```mermaid
sequenceDiagram
    participant Main
    participant EDA
    participant Consumer
    participant RabbitMQ
    
    Main->>EDA: NewEDA(url)
    Main->>EDA: AddConsumer(projectConsumer)
    Main->>EDA: AddConsumer(permissionConsumer)
    Main->>EDA: StartConsumers()
    
    loop Para cada Consumer
        EDA->>RabbitMQ: Connect
        RabbitMQ-->>EDA: Connection
        EDA->>RabbitMQ: Create Channel
        RabbitMQ-->>EDA: Channel
        EDA->>Consumer: Declare(channel)
        Consumer->>RabbitMQ: ExchangeDeclarePassive
        Consumer->>RabbitMQ: QueueDeclarePassive
        Consumer->>RabbitMQ: QueueBind
        EDA->>Consumer: Consume(channel)
        Consumer->>RabbitMQ: Start Consuming
    end
```

## Código de Inicialização

```go
// No arquivo codeactions.go
if cfg.EDA.RabbitmqURL != "" {
    eda := rabbitmq.NewEDA(cfg.EDA.RabbitmqURL)

    // Serviços de permissão e projeto
    permissionService := permission.NewUserPermissionService(repo)
    projectService := project.NewProjectService(repo)

    // Criar consumers
    projectConsumer := project.NewProjectConsumer(
        projectService,
        permissionService,
        cfg.EDA.ProjectExchangeName,
        cfg.EDA.ProjectQueueName,
    )
    
    permissionConsumer := permission.NewPermissionConsumer(
        permissionService,
        cfg.EDA.PermissionExchangeName,
        cfg.EDA.PermissionQueueName,
    )

    // Registrar consumers
    eda.AddConsumer(projectConsumer)
    eda.AddConsumer(permissionConsumer)

    // Iniciar consumo
    if err := eda.StartConsumers(); err != nil {
        log.WithError(err)
    }
}
```

## Diagrama de Sequência Completo

```mermaid
sequenceDiagram
    participant External as Sistema Externo
    participant RMQ as RabbitMQ
    participant EDA as EDA Manager
    participant PC as Project Consumer
    participant PS as Project Service
    participant PMS as Permission Service
    participant DB as Database
    
    Note over External,DB: Criação de Novo Projeto
    
    External->>RMQ: Publish Project Event
    RMQ->>PC: Deliver Message
    PC->>PC: Parse JSON
    PC->>PS: FindByUUID(uuid)
    PS->>DB: SELECT
    DB-->>PS: Not Found
    PS-->>PC: nil
    
    PC->>PS: Create(project)
    PS->>DB: INSERT project
    DB-->>PS: OK
    
    loop Para cada Authorization
        PC->>PMS: Create(permission)
        PMS->>DB: INSERT permission
        DB-->>PMS: OK
    end
    
    PC->>RMQ: ACK
    
    Note over External,DB: Atualização de Permissão
    
    External->>RMQ: Publish Permission Event
    RMQ->>PC: Deliver Message
    PC->>PC: Parse JSON
    PC->>PMS: Find(project_uuid, email)
    PMS->>DB: SELECT
    DB-->>PMS: Found
    
    PC->>PMS: Update(permission)
    PMS->>DB: UPDATE
    DB-->>PMS: OK
    
    PC->>RMQ: ACK
```

## Melhores Práticas

1. **Idempotência**: Os handlers são projetados para serem idempotentes
2. **Validação**: Mensagens inválidas são descartadas com ACK
3. **Logs**: Todas as operações são logadas para debugging
4. **Reconexão**: O sistema tenta reconectar em caso de falha
5. **QoS**: Prefetch de 1 mensagem por vez para processamento ordenado
