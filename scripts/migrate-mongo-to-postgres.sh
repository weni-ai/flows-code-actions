#!/bin/bash
#
# Script completo de migração: MongoDB → CSV → PostgreSQL
# Este script automatiza todo o processo de migração
#
# Uso: ./migrate-mongo-to-postgres.sh [opcoes]
#

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Funções de utilidade
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================${NC}"
}

# Parse de argumentos
DRY_RUN=false
SKIP_EXPORT=false
SKIP_IMPORT=false
BATCH_SIZE=1000
CSV_DIR=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-export)
            SKIP_EXPORT=true
            shift
            ;;
        --skip-import)
            SKIP_IMPORT=true
            shift
            ;;
        --batch-size)
            BATCH_SIZE="$2"
            shift 2
            ;;
        --csv-dir)
            CSV_DIR="$2"
            shift 2
            ;;
        -h|--help)
            echo "Uso: $0 [opcoes]"
            echo ""
            echo "Opções:"
            echo "  --dry-run          Executar em modo de teste (sem alterações)"
            echo "  --skip-export      Pular etapa de exportação do MongoDB"
            echo "  --skip-import      Pular etapa de importação no PostgreSQL"
            echo "  --csv-dir DIR      Usar diretório CSV específico"
            echo "  --batch-size N     Tamanho do batch para importação (padrão: 1000)"
            echo "  -h, --help         Mostrar esta mensagem"
            echo ""
            echo "Exemplos:"
            echo "  $0                                    # Migração completa"
            echo "  $0 --dry-run                          # Teste sem alterações"
            echo "  $0 --skip-export --csv-dir exports/   # Apenas importar CSVs existentes"
            echo "  $0 --batch-size 5000                  # Com batch customizado"
            exit 0
            ;;
        *)
            log_error "Opção desconhecida: $1"
            exit 1
            ;;
    esac
done

# Diretório do script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

print_header "Migração MongoDB → PostgreSQL"
echo ""

if [ "$DRY_RUN" = true ]; then
    log_warning "MODO DRY-RUN ATIVADO - Nenhuma alteração será feita"
    echo ""
fi

# Verificar dependências
log_info "Verificando dependências..."

if [ "$SKIP_EXPORT" = false ]; then
    if ! command -v mongoexport &> /dev/null; then
        log_error "mongoexport não encontrado!"
        echo "Instale o MongoDB Database Tools:"
        echo "  Ubuntu/Debian: sudo apt-get install mongodb-database-tools"
        echo "  MacOS: brew install mongodb-database-tools"
        echo "  Fedora: sudo dnf install mongodb-database-tools"
        exit 1
    fi
    log_success "mongoexport encontrado"
fi

if [ "$SKIP_IMPORT" = false ]; then
    if ! command -v go &> /dev/null; then
        log_error "Go não encontrado!"
        echo "Instale o Go: https://golang.org/doc/install"
        exit 1
    fi
    log_success "Go encontrado ($(go version))"
fi

echo ""

# Etapa 1: Exportar do MongoDB
if [ "$SKIP_EXPORT" = false ]; then
    print_header "Etapa 1: Exportando dados do MongoDB"
    echo ""
    
    if [ -f "$SCRIPT_DIR/export-mongo-to-csv.sh" ]; then
        bash "$SCRIPT_DIR/export-mongo-to-csv.sh"
        
        # Pegar o diretório mais recente
        if [ -z "$CSV_DIR" ]; then
            CSV_DIR=$(ls -td "$PROJECT_ROOT/mongo_exports"/*/ 2>/dev/null | head -1)
            if [ -z "$CSV_DIR" ]; then
                log_error "Nenhum diretório de exportação encontrado"
                exit 1
            fi
            CSV_DIR="${CSV_DIR%/}" # Remover trailing slash
        fi
        
        log_success "Exportação concluída: $CSV_DIR"
    else
        log_error "Script de exportação não encontrado: $SCRIPT_DIR/export-mongo-to-csv.sh"
        exit 1
    fi
else
    log_info "Pulando exportação (--skip-export)"
    
    if [ -z "$CSV_DIR" ]; then
        log_error "Você deve especificar --csv-dir quando usar --skip-export"
        exit 1
    fi
    
    if [ ! -d "$CSV_DIR" ]; then
        log_error "Diretório CSV não encontrado: $CSV_DIR"
        exit 1
    fi
fi

echo ""

# Etapa 2: Importar para PostgreSQL
if [ "$SKIP_IMPORT" = false ]; then
    print_header "Etapa 2: Importando dados para PostgreSQL"
    echo ""
    
    if [ -f "$SCRIPT_DIR/import-csv-to-postgres.go" ]; then
        # Construir argumentos
        IMPORT_ARGS="-dir=$CSV_DIR -batch-size=$BATCH_SIZE"
        
        if [ "$DRY_RUN" = true ]; then
            IMPORT_ARGS="$IMPORT_ARGS -dry-run"
        fi
        
        log_info "Executando importador..."
        cd "$PROJECT_ROOT"
        go run "$SCRIPT_DIR/import-csv-to-postgres.go" $IMPORT_ARGS
        
        log_success "Importação concluída!"
    else
        log_error "Script de importação não encontrado: $SCRIPT_DIR/import-csv-to-postgres.go"
        exit 1
    fi
else
    log_info "Pulando importação (--skip-import)"
fi

echo ""

# Resumo final
print_header "Migração Concluída"
echo ""

if [ "$DRY_RUN" = false ]; then
    log_success "Dados migrados com sucesso!"
    echo ""
    log_info "Próximos passos:"
    echo "  1. Verificar os dados no PostgreSQL"
    echo "  2. Testar a aplicação com PostgreSQL"
    echo "  3. Ajustar FLOWS_CODE_ACTIONS_DB_TYPE=postgres"
    echo ""
    log_info "Comandos úteis:"
    echo "  # Contar registros"
    echo "  psql \$FLOWS_CODE_ACTIONS_DB_URI -c 'SELECT COUNT(*) FROM codes;'"
    echo "  psql \$FLOWS_CODE_ACTIONS_DB_URI -c 'SELECT COUNT(*) FROM codelibs;'"
    echo "  psql \$FLOWS_CODE_ACTIONS_DB_URI -c 'SELECT COUNT(*) FROM coderuns;'"
    echo ""
    echo "  # Ver últimos registros"
    echo "  psql \$FLOWS_CODE_ACTIONS_DB_URI -c 'SELECT id, name, type, created_at FROM codes ORDER BY created_at DESC LIMIT 5;'"
else
    log_info "Modo dry-run concluído. Nenhuma alteração foi feita."
    echo ""
    log_info "Para executar a migração real, rode:"
    echo "  $0"
fi

echo ""

