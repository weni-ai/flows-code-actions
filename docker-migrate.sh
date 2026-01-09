#!/bin/bash
#
# Helper script para Docker standalone (sem docker-compose)
# Para quando você já tem MongoDB, PostgreSQL e Redis rodando
#
# Uso: ./docker-standalone.sh [comando]
#

set -e

IMAGE_NAME="codeactions:latest"
CONTAINER_NAME="codeactions-app"
ENV_FILE=".env.docker"

# Cores
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_help() {
    echo ""
    echo "════════════════════════════════════════════════════════════"
    echo "  Code Actions - Docker Standalone (Sem Docker Compose)"
    echo "════════════════════════════════════════════════════════════"
    echo ""
    echo "Uso: $0 [comando]"
    echo ""
    echo "Setup:"
    echo "  build           - Build da imagem Docker"
    echo "  env             - Criar arquivo .env.docker de exemplo"
    echo ""
    echo "Migração:"
    echo "  migrate                      - Criar tabelas no PostgreSQL"
    echo "  export                       - Exportar dados do MongoDB"
    echo "  import [dir]                 - Importar CSVs para PostgreSQL"
    echo "  import-collection <coll> [dir] - Importar apenas uma collection"
    echo "  list-collections             - Listar collections disponíveis"
    echo "  full-migration               - Migração completa (export + import)"
    echo "  verify                       - Verificar dados importados"
    echo ""
    echo "Aplicação:"
    echo "  start           - Iniciar aplicação (daemon)"
    echo "  stop            - Parar aplicação"
    echo "  restart         - Reiniciar aplicação"
    echo "  logs            - Ver logs da aplicação"
    echo "  status          - Ver status do container"
    echo ""
    echo "Debug:"
    echo "  shell           - Entrar no container (bash)"
    echo "  test-mongo      - Testar conexão com MongoDB"
    echo "  test-postgres   - Testar conexão com PostgreSQL"
    echo ""
    echo "Exemplos:"
    echo "  $0 build                              # Build da imagem"
    echo "  $0 env                                # Criar .env.docker"
    echo "  $0 migrate                            # Criar tabelas"
    echo "  $0 full-migration                     # Migrar dados (tudo)"
    echo "  $0 import-collection code             # Importar apenas 'code'"
    echo "  $0 import-collection coderun ./exports # Importar 'coderun' de dir específico"
    echo "  $0 start                              # Iniciar app"
    echo ""
}

ensure_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${YELLOW}⚠  Arquivo $ENV_FILE não encontrado${NC}"
        echo "Crie com: $0 env"
        exit 1
    fi
}

case "$1" in
    build)
        echo -e "${BLUE}🔨 Building imagem Docker...${NC}"
        docker build -t "$IMAGE_NAME" .
        echo ""
        echo -e "${GREEN}✓ Build concluído${NC}"
        echo "Imagem: $IMAGE_NAME"
        ;;
    
    env)
        if [ -f "$ENV_FILE" ]; then
            echo -e "${YELLOW}⚠  Arquivo $ENV_FILE já existe${NC}"
            echo "Deseja sobrescrever? (y/N)"
            read -r response
            if [[ ! "$response" =~ ^[Yy]$ ]]; then
                echo "Cancelado"
                exit 0
            fi
        fi
        
        echo -e "${BLUE}📝 Criando $ENV_FILE...${NC}"
        cat > "$ENV_FILE" << 'EOF'
# MongoDB (origem dos dados)
FLOWS_CODE_ACTIONS_MONGO_DB_URI=mongodb://localhost:27017
FLOWS_CODE_ACTIONS_MONGO_DB_NAME=code-actions

# PostgreSQL (destino e banco principal)
FLOWS_CODE_ACTIONS_DB_TYPE=postgres
FLOWS_CODE_ACTIONS_DB_URI=postgres://postgres:postgres@localhost:5432/codeactions?sslmode=disable
FLOWS_CODE_ACTIONS_DB_NAME=codeactions

# Redis
FLOWS_CODE_ACTIONS_REDIS=redis://localhost:6379/10

# Aplicação
FLOWS_CODE_ACTIONS_ENVIRONMENT=production
FLOWS_CODE_ACTIONS_LOG_LEVEL=info
FLOWS_CODE_ACTIONS_HOST=":"
FLOWS_CODE_ACTIONS_PORT=8080
EOF
        
        echo -e "${GREEN}✓ Arquivo $ENV_FILE criado${NC}"
        echo ""
        echo "Edite o arquivo para ajustar as URIs conforme seu ambiente:"
        echo "  nano $ENV_FILE"
        echo "  ou"
        echo "  code $ENV_FILE"
        ;;
    
    migrate)
        ensure_env_file
        echo -e "${BLUE}📊 Criando tabelas no PostgreSQL...${NC}"
        
        # Source env file para obter DB_URI
        source "$ENV_FILE"
        
        docker run --rm \
            --network host \
            --env-file "$ENV_FILE" \
            "$IMAGE_NAME" \
            /app/migrate -path /app/migrations \
            -database "$FLOWS_CODE_ACTIONS_DB_URI" up
        
        echo ""
        echo -e "${GREEN}✓ Migrations executadas${NC}"
        ;;
    
    export)
        ensure_env_file
        echo -e "${BLUE}📤 Exportando dados do MongoDB...${NC}"
        
        mkdir -p ./mongo_exports
        
        docker run --rm \
            --network host \
            --env-file "$ENV_FILE" \
            -v "$(pwd)/mongo_exports:/app/mongo_exports" \
            "$IMAGE_NAME" \
            bash /app/scripts/export-mongo-to-csv.sh
        
        echo ""
        echo -e "${GREEN}✓ Exportação concluída${NC}"
        echo "CSVs salvos em: ./mongo_exports/"
        ;;
    
    import)
        ensure_env_file
        
        if [ -z "$2" ]; then
            echo -e "${YELLOW}⚠  Especifique o diretório dos CSVs${NC}"
            echo "Uso: $0 import [diretório]"
            echo ""
            echo "Diretórios disponíveis:"
            ls -1d mongo_exports/*/ 2>/dev/null || echo "  (nenhum)"
            exit 1
        fi
        
        echo -e "${BLUE}📥 Importando CSVs para PostgreSQL...${NC}"
        
        docker run --rm \
            --network host \
            --env-file "$ENV_FILE" \
            -v "$(pwd)/mongo_exports:/app/mongo_exports" \
            "$IMAGE_NAME" \
            go run /app/scripts/import-csv-to-postgres.go -dir="$2"
        
        echo ""
        echo -e "${GREEN}✓ Importação concluída${NC}"
        ;;
    
    list-collections)
        echo -e "${BLUE}📋 Collections disponíveis:${NC}"
        echo ""
        docker run --rm "$IMAGE_NAME" \
            go run /app/scripts/import-csv-to-postgres.go -list-collections
        ;;
    
    import-collection)
        ensure_env_file
        
        if [ -z "$2" ]; then
            echo -e "${RED}✗ Erro: Especifique a collection${NC}"
            echo ""
            echo "Uso: $0 import-collection <collection> [diretório]"
            echo ""
            echo "Collections disponíveis:"
            echo "  - code"
            echo "  - codelib"
            echo "  - coderun"
            echo "  - projects"
            echo "  - user_permissions"
            echo ""
            echo "Exemplos:"
            echo "  $0 import-collection code"
            echo "  $0 import-collection coderun ./mongo_exports/20231230_143000"
            exit 1
        fi
        
        COLLECTION="$2"
        
        # Determinar diretório
        if [ -n "$3" ]; then
            CSV_DIR="$3"
        else
            # Pegar o diretório mais recente
            CSV_DIR=$(ls -1td mongo_exports/*/ 2>/dev/null | head -n1)
            if [ -z "$CSV_DIR" ]; then
                echo -e "${RED}✗ Nenhum diretório de exportação encontrado${NC}"
                echo ""
                echo "Execute primeiro: $0 export"
                exit 1
            fi
            CSV_DIR="${CSV_DIR%/}" # Remove trailing slash
            echo -e "${YELLOW}ℹ  Usando diretório mais recente: $CSV_DIR${NC}"
        fi
        
        echo -e "${BLUE}📥 Importando collection: ${YELLOW}$COLLECTION${NC}"
        echo -e "${BLUE}   Diretório: $CSV_DIR${NC}"
        echo ""
        
        docker run --rm \
            --network host \
            --env-file "$ENV_FILE" \
            -v "$(pwd)/mongo_exports:/app/mongo_exports" \
            "$IMAGE_NAME" \
            go run /app/scripts/import-csv-to-postgres.go \
                -dir="$CSV_DIR" \
                -collection="$COLLECTION"
        
        echo ""
        echo -e "${GREEN}✓ Collection $COLLECTION importada${NC}"
        ;;
    
    full-migration)
        ensure_env_file
        echo -e "${BLUE}🔄 Executando migração completa...${NC}"
        
        mkdir -p ./mongo_exports
        
        docker run --rm \
            --network host \
            --env-file "$ENV_FILE" \
            -v "$(pwd)/mongo_exports:/app/mongo_exports" \
            "$IMAGE_NAME" \
            bash /app/scripts/migrate-mongo-to-postgres.sh
        
        echo ""
        echo -e "${GREEN}✓ Migração completa concluída${NC}"
        ;;
    
    verify)
        ensure_env_file
        echo -e "${BLUE}🔍 Verificando dados...${NC}"
        
        source "$ENV_FILE"
        
        docker run --rm \
            --network host \
            --env-file "$ENV_FILE" \
            "$IMAGE_NAME" \
            psql "$FLOWS_CODE_ACTIONS_DB_URI" -c "
                SELECT 'codes' as table_name, COUNT(*) as total FROM codes
                UNION ALL
                SELECT 'codelibs', COUNT(*) FROM codelibs
                UNION ALL
                SELECT 'coderuns', COUNT(*) FROM coderuns;
            "
        ;;
    
    start)
        ensure_env_file
        
        # Verificar se já está rodando
        if docker ps | grep -q "$CONTAINER_NAME"; then
            echo -e "${YELLOW}⚠  Container já está rodando${NC}"
            echo "Use: $0 restart"
            exit 0
        fi
        
        echo -e "${BLUE}🚀 Iniciando aplicação...${NC}"
        
        docker run -d \
            --name "$CONTAINER_NAME" \
            --network host \
            --env-file "$ENV_FILE" \
            -v "$(pwd)/codes:/app/codes" \
            --restart unless-stopped \
            "$IMAGE_NAME"
        
        echo ""
        echo -e "${GREEN}✓ Aplicação iniciada${NC}"
        echo "Container: $CONTAINER_NAME"
        echo "API: http://localhost:8080"
        echo ""
        echo "Ver logs: $0 logs"
        ;;
    
    stop)
        echo -e "${BLUE}🛑 Parando aplicação...${NC}"
        docker stop "$CONTAINER_NAME"
        docker rm "$CONTAINER_NAME"
        echo -e "${GREEN}✓ Aplicação parada${NC}"
        ;;
    
    restart)
        echo -e "${BLUE}🔄 Reiniciando aplicação...${NC}"
        docker restart "$CONTAINER_NAME"
        echo -e "${GREEN}✓ Aplicação reiniciada${NC}"
        ;;
    
    logs)
        echo -e "${BLUE}📋 Logs da aplicação (Ctrl+C para sair):${NC}"
        docker logs -f --tail 100 "$CONTAINER_NAME"
        ;;
    
    status)
        echo -e "${BLUE}📊 Status do container:${NC}"
        docker ps -a | grep "$CONTAINER_NAME" || echo "Container não encontrado"
        echo ""
        echo "Estatísticas de uso:"
        docker stats --no-stream "$CONTAINER_NAME" 2>/dev/null || echo "Container não está rodando"
        ;;
    
    shell)
        ensure_env_file
        echo -e "${BLUE}🐚 Abrindo shell no container...${NC}"
        docker run --rm -it \
            --network host \
            --env-file "$ENV_FILE" \
            -v "$(pwd)/mongo_exports:/app/mongo_exports" \
            --entrypoint bash \
            "$IMAGE_NAME"
        ;;
    
    test-mongo)
        ensure_env_file
        echo -e "${BLUE}🧪 Testando conexão com MongoDB...${NC}"
        
        source "$ENV_FILE"
        
        docker run --rm \
            --network host \
            mongo:7 \
            mongosh "$FLOWS_CODE_ACTIONS_MONGO_DB_URI" \
            --eval "db.adminCommand('ping')"
        
        echo ""
        echo -e "${GREEN}✓ MongoDB acessível${NC}"
        ;;
    
    test-postgres)
        ensure_env_file
        echo -e "${BLUE}🧪 Testando conexão com PostgreSQL...${NC}"
        
        source "$ENV_FILE"
        
        docker run --rm \
            --network host \
            postgres:16-alpine \
            psql "$FLOWS_CODE_ACTIONS_DB_URI" -c "SELECT 1;"
        
        echo ""
        echo -e "${GREEN}✓ PostgreSQL acessível${NC}"
        ;;
    
    *)
        print_help
        exit 0
        ;;
esac

