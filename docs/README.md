# Flows Code Actions - Documentação Completa

## Índice

1. [Visão Geral](./01-visao-geral.md)
2. [Arquitetura](./02-arquitetura.md)
3. [Entidades e Modelos](./03-entidades.md)
4. [Endpoints da API](./04-endpoints.md)
5. [Migrações de Banco de Dados](./05-migracoes.md)
6. [Sistema de Permissões](./06-permissoes.md)
7. [Event-Driven Architecture (EDA)](./07-eda.md)
8. [Configuração](./08-configuracao.md)

---

## Sobre o Projeto

O **Flows Code Actions** é uma API desenvolvida em Go que permite a execução dinâmica de código em múltiplas linguagens (Python, JavaScript, Go). O sistema suporta dois tipos principais de código:

- **Flow**: Código executado internamente como parte de um fluxo de automação
- **Endpoint**: Código exposto como endpoint HTTP acessível externamente

### Tecnologias Principais

| Tecnologia | Uso |
|------------|-----|
| **Go 1.21+** | Linguagem principal do backend |
| **Echo** | Framework HTTP |
| **PostgreSQL** | Banco de dados principal |
| **MongoDB** | Banco de dados legado (suporte mantido) |
| **Redis** | Cache, rate limiting e locks distribuídos |
| **RabbitMQ** | Mensageria para eventos (projetos e permissões) |
| **Amazon S3** | Armazenamento de logs de execução |
| **Docker** | Containerização |

### Estrutura de Diretórios

```
flows-code-actions/
├── cmd/codeactions/          # Ponto de entrada da aplicação
├── config/                   # Configurações
├── internal/                 # Código interno da aplicação
│   ├── code/                 # Gerenciamento de códigos
│   ├── codelib/              # Bibliotecas de código
│   ├── codelog/              # Logs de execução
│   ├── coderun/              # Execuções de código
│   ├── coderunner/           # Engine de execução
│   ├── db/                   # Conexões com banco de dados
│   ├── eventdriven/          # Event-driven architecture
│   ├── http/echo/            # Handlers e rotas HTTP
│   ├── metrics/              # Métricas Prometheus
│   ├── permission/           # Sistema de permissões
│   └── project/              # Gerenciamento de projetos
├── migrations/               # Migrações SQL
├── engines/                  # Engines de execução por linguagem
│   ├── go/                   # Engine Go
│   └── py/                   # Engine Python
├── scripts/                  # Scripts utilitários
└── doc/                      # Documentação adicional
```

### Links Rápidos

- [Como executar o projeto](../README.md)
- [Postman Collection](../doc/code_actions_postman_collection_v4.json)
- [Configuração do LocalStack](../README.md#localstack-s3-development-setup)
