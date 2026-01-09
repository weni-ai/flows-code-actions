# Scripts de Migração MongoDB → PostgreSQL

Este diretório contém scripts para exportar dados do MongoDB para CSV e importar para o PostgreSQL.

## Estrutura

- `export-mongo-to-csv.sh` - Script bash para exportar collections do MongoDB para CSV
- `import-csv-to-postgres.go` - Programa Go para importar CSVs no PostgreSQL

## Pré-requisitos

### Para exportação (MongoDB)
```bash
# Instalar MongoDB Database Tools
# Ubuntu/Debian
sudo apt-get install mongodb-database-tools

# MacOS
brew install mongodb-database-tools

# Fedora
sudo dnf install mongodb-database-tools
```

### Para importação (PostgreSQL)
```bash
# O programa Go já possui as dependências necessárias no go.mod
# Apenas certifique-se de ter o Go instalado
go version
```

## Uso

### 1. Exportar dados do MongoDB para CSV

```bash
# Dar permissão de execução ao script
chmod +x scripts/export-mongo-to-csv.sh

# Executar com configurações padrão (variáveis de ambiente)
./scripts/export-mongo-to-csv.sh

# OU definir manualmente as configurações
FLOWS_CODE_ACTIONS_MONGO_DB_URI="mongodb://user:pass@localhost:27017" \
FLOWS_CODE_ACTIONS_MONGO_DB_NAME="code-actions" \
./scripts/export-mongo-to-csv.sh
```

**Saída:**
- Os arquivos CSV serão salvos em: `./mongo_exports/YYYYMMDD_HHMMSS/`
- Cada collection terá seu próprio arquivo CSV

**Collections exportadas:**
- `code.csv` - Códigos/actions (flows e endpoints)
- `codelib.csv` - Bibliotecas de código
- `coderun.csv` - Execuções de código
- `projects.csv` - Projetos (se existir)
- `user_permissions.csv` - Permissões de usuário (se existir)

### 2. Importar CSVs para o PostgreSQL

```bash
# Importação básica
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231201_120000

# Com URI personalizada do PostgreSQL
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231201_120000 \
  -pg-uri="postgres://user:pass@localhost:5432/codeactions?sslmode=disable"

# Modo dry-run (não faz alterações, apenas testa)
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231201_120000 \
  -dry-run

# Ignorar erros e continuar
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231201_120000 \
  -skip-errors

# Ajustar tamanho do batch
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231201_120000 \
  -batch-size=500
```

### 3. Compilar o importador (opcional)

```bash
# Compilar binário
go build -o bin/import-csv-to-postgres scripts/import-csv-to-postgres.go

# Executar binário compilado
./bin/import-csv-to-postgres -dir=./mongo_exports/20231201_120000
```

## Opções do Importador

| Flag | Descrição | Padrão |
|------|-----------|--------|
| `-dir` | Diretório contendo os CSVs | (obrigatório) |
| `-pg-uri` | URI de conexão do PostgreSQL | Usa variável de ambiente `FLOWS_CODE_ACTIONS_DB_URI` |
| `-dry-run` | Executar sem fazer alterações | `false` |
| `-batch-size` | Tamanho do lote para inserções | `1000` |
| `-skip-errors` | Continuar mesmo se houver erros | `false` |

## Fluxo Completo de Migração

```bash
# 1. Exportar do MongoDB
./scripts/export-mongo-to-csv.sh

# 2. Verificar arquivos exportados
ls -lh mongo_exports/20231201_120000/

# 3. Testar importação (dry-run)
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231201_120000 \
  -dry-run

# 4. Importar para PostgreSQL
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231201_120000

# 5. Verificar dados no PostgreSQL
psql -h localhost -U postgres -d codeactions -c "SELECT COUNT(*) FROM codes;"
psql -h localhost -U postgres -d codeactions -c "SELECT COUNT(*) FROM codelibs;"
psql -h localhost -U postgres -d codeactions -c "SELECT COUNT(*) FROM coderuns;"
```

## Configuração de Variáveis de Ambiente

### MongoDB
```bash
export FLOWS_CODE_ACTIONS_MONGO_DB_URI="mongodb://localhost:27017"
export FLOWS_CODE_ACTIONS_MONGO_DB_NAME="code-actions"
```

### PostgreSQL
```bash
export FLOWS_CODE_ACTIONS_DB_URI="postgres://localhost:5432/codeactions?sslmode=disable"
export FLOWS_CODE_ACTIONS_DB_NAME="codeactions"
```

## Recursos e Características

### Script de Exportação
- ✅ Exporta todas as collections relevantes
- ✅ Cria timestamp automático para organização
- ✅ Mostra progresso e contagem de registros
- ✅ Trata collections opcionais gracefully

### Programa de Importação
- ✅ Importação em batch para performance
- ✅ Transações para garantir consistência
- ✅ Modo dry-run para testes
- ✅ Tratamento de erros configurável
- ✅ Suporte a ON CONFLICT (upsert)
- ✅ Parse de JSON para campos JSONB
- ✅ Parse de datas em múltiplos formatos
- ✅ Relacionamento entre tabelas (code_id em coderuns)
- ✅ Estatísticas detalhadas de importação

## Tratamento de Dados

### IDs
- O `_id` do MongoDB é salvo no campo `mongo_object_id` no PostgreSQL
- O PostgreSQL gera novos UUIDs automaticamente como chaves primárias

### Campos JSON
- Campos `extra`, `params` e `headers` são convertidos para JSONB
- Valores vazios são convertidos para `{}` (objeto JSON vazio)

### Datas
- Suporte a múltiplos formatos: ISO 8601, RFC3339, etc.
- Fallback para data atual em caso de erro

### Relacionamentos
- Em `coderuns`, o campo `code_id` é resolvido automaticamente
- Se o código referenciado não existir, o campo fica NULL

## Troubleshooting

### Erro: mongoexport não encontrado
```bash
# Instalar MongoDB Database Tools
sudo apt-get install mongodb-database-tools
```

### Erro: conexão recusada no MongoDB
```bash
# Verificar se o MongoDB está rodando
sudo systemctl status mongod

# Verificar URI de conexão
echo $FLOWS_CODE_ACTIONS_MONGO_DB_URI
```

### Erro: conexão recusada no PostgreSQL
```bash
# Verificar se o PostgreSQL está rodando
sudo systemctl status postgresql

# Testar conexão
psql -h localhost -U postgres -d codeactions -c "SELECT 1;"
```

### Erro: collection não encontrada
- Isso é normal para collections opcionais (`projects`, `user_permissions`)
- O script continua normalmente

### Importação muito lenta
```bash
# Aumentar o tamanho do batch
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/... -batch-size=5000
```

## Exemplo de Saída

### Exportação
```
============================================
Exportando dados do MongoDB para CSV
============================================
MongoDB URI: mongodb://localhost:27017
Database: code-actions
Diretório de saída: ./mongo_exports/20231201_120000
============================================

Exportando collection: code...
✓ code exportado com sucesso (1523 registros)

Exportando collection: codelib...
✓ codelib exportado com sucesso (45 registros)

Exportando collection: coderun...
✓ coderun exportado com sucesso (8932 registros)

============================================
Exportação concluída!
============================================
```

### Importação
```
============================================
Importação de CSV para PostgreSQL
============================================
Diretório CSV: ./mongo_exports/20231201_120000
PostgreSQL URI: postgres://localhost:5432/codeactions
============================================

✓ Conectado ao PostgreSQL com sucesso

Importando: code
✓ code importado: 1523 total, 1523 sucesso, 0 falhas, 0 pulados (2.34s)

Importando: codelib
✓ codelib importado: 45 total, 45 sucesso, 0 falhas, 0 pulados (0.12s)

Importando: coderun
✓ coderun importado: 8932 total, 8932 sucesso, 0 falhas, 0 pulados (12.45s)

============================================
Resumo da Importação
============================================
code: 1523/1523 registros importados
codelib: 45/45 registros importados
coderun: 8932/8932 registros importados
============================================
```

## Notas Importantes

⚠️ **Backup**: Sempre faça backup dos seus dados antes de executar a migração

⚠️ **Dry-run**: Execute primeiro em modo `-dry-run` para verificar se tudo está correto

⚠️ **Ordem**: As collections são importadas em ordem para respeitar dependências

⚠️ **Upsert**: O importador usa `ON CONFLICT DO UPDATE`, então pode ser executado múltiplas vezes

⚠️ **Performance**: Para grandes volumes, ajuste o `-batch-size` conforme necessário

