# Migração MongoDB → PostgreSQL - Passo a Passo

## ⚡ Migração Rápida (Recomendado)

### 1. Configure as variáveis de ambiente

```bash
export FLOWS_CODE_ACTIONS_MONGO_DB_URI="mongodb://localhost:27017"
export FLOWS_CODE_ACTIONS_MONGO_DB_NAME="code-actions"
export FLOWS_CODE_ACTIONS_DB_URI="postgres://user:pass@localhost:5432/codeactions?sslmode=disable"
```

### 2. Execute a migração completa

```bash
./scripts/migrate-data.sh
```

**Pronto!** O script faz tudo automaticamente:
- ✅ Valida conexões
- ✅ Mostra estatísticas ANTES
- ✅ Exporta do MongoDB
- ✅ Importa para PostgreSQL
- ✅ Mostra estatísticas DEPOIS
- ✅ Verifica integridade

---

## 🔧 Migração Manual (Passo a Passo)

### 1. Criar tabelas no PostgreSQL

```bash
# Baixar migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Executar migrations
migrate -path ./migrations -database "$FLOWS_CODE_ACTIONS_DB_URI" up
```

### 2. Exportar dados do MongoDB

```bash
./scripts/export-mongo-to-csv.sh
```

**Saída:** CSVs em `./mongo_exports/YYYYMMDD_HHMMSS/`

### 3. Importar para PostgreSQL

```bash
# Importar tudo
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231230_143000

# OU importar uma collection de cada vez
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231230_143000 -collection=code
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231230_143000 -collection=codelib
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231230_143000 -collection=projects
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231230_143000 -collection=coderun
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231230_143000 -collection=user_permissions
```

### 4. Verificar dados

```bash
psql "$FLOWS_CODE_ACTIONS_DB_URI" << 'SQL'
SELECT 'codes' as tabela, COUNT(*) as total FROM codes
UNION ALL
SELECT 'codelibs', COUNT(*) FROM codelibs
UNION ALL
SELECT 'coderuns', COUNT(*) FROM coderuns
UNION ALL
SELECT 'projects', COUNT(*) FROM projects
UNION ALL
SELECT 'user_permissions', COUNT(*) FROM user_permissions;
SQL
```

---

## 🐳 Migração com Docker

### 1. Configurar ambiente

```bash
# Criar .env.docker
cat > .env.docker << 'EOF'
FLOWS_CODE_ACTIONS_MONGO_DB_URI=mongodb://host.docker.internal:27017
FLOWS_CODE_ACTIONS_MONGO_DB_NAME=code-actions
FLOWS_CODE_ACTIONS_DB_TYPE=postgres
FLOWS_CODE_ACTIONS_DB_URI=postgres://postgres:postgres@host.docker.internal:5432/codeactions?sslmode=disable
FLOWS_CODE_ACTIONS_DB_NAME=codeactions
FLOWS_CODE_ACTIONS_REDIS=redis://host.docker.internal:6379/10
EOF

# Ajustar valores conforme seu ambiente
nano .env.docker
```

### 2. Build da imagem

```bash
docker build -t codeactions:latest .
```

### 3. Criar tabelas

```bash
docker run --rm --network host --env-file .env.docker \
  codeactions:latest \
  /app/migrate -path /app/migrations -database "$FLOWS_CODE_ACTIONS_DB_URI" up
```

### 4. Migrar dados

```bash
# Migração completa
docker run --rm --network host --env-file .env.docker \
  -v $(pwd)/mongo_exports:/app/mongo_exports \
  codeactions:latest \
  bash /app/scripts/migrate-data.sh

# OU usando helper script
./docker-migrate.sh full-migration
```

### 5. Iniciar aplicação

```bash
docker run -d --name codeactions-app \
  --network host \
  --env-file .env.docker \
  codeactions:latest
```

---

## 📊 Opções Avançadas

### Ajustar batch-size (para evitar OOM)

```bash
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231230_143000 \
  -batch-size=50
```

### Importar com verbose (ver detalhes)

```bash
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231230_143000 \
  -collection=coderun \
  -verbose
```

### Modo dry-run (testar sem importar)

```bash
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231230_143000 \
  -dry-run
```

### Continuar mesmo com erros

```bash
go run scripts/import-csv-to-postgres.go \
  -dir=./mongo_exports/20231230_143000 \
  -skip-errors
```

---

## ❌ Solução de Problemas

### "Killed" após importação

```bash
# Usar batch menor
go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/... -batch-size=50

# Ver documentação
cat scripts/KILLED_ISSUE.md
```

### Erro de conexão PostgreSQL

```bash
# Testar conexão
psql "$FLOWS_CODE_ACTIONS_DB_URI" -c "SELECT 1"
```

### Erro de conexão MongoDB

```bash
# Testar conexão
mongosh "$FLOWS_CODE_ACTIONS_MONGO_DB_URI/$FLOWS_CODE_ACTIONS_MONGO_DB_NAME" --eval "db.stats()"
```

### Arquivo CSV vazio

```bash
# Script pula automaticamente, mas você pode verificar:
ls -lh mongo_exports/*/
```

---

## 📚 Documentação Completa

- **Guia Completo:** `MIGRATION_GUIDE.md`
- **Docker:** `DOCKER_MIGRATION.md`
- **Importação Individual:** `scripts/IMPORT_INDIVIDUAL.md`
- **Troubleshooting:** `scripts/TROUBLESHOOTING.md`
- **Problema "Killed":** `scripts/KILLED_ISSUE.md`

---

## ⏱️ Tempo Estimado

| Dataset | Tempo Aproximado |
|---------|------------------|
| < 1K registros | 1-2 minutos |
| 1K - 10K | 5-10 minutos |
| 10K - 100K | 15-30 minutos |
| 100K - 1M | 1-3 horas |
| > 1M | 3+ horas |

**Dica:** Use `-batch-size=100` (padrão) para melhor balanço entre velocidade e memória.

