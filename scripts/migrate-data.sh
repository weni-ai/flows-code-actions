#!/bin/bash
#
# Script completo de migração MongoDB → PostgreSQL
# Exporta do MongoDB para CSV e importa no PostgreSQL
#
# Uso: ./migrate-data.sh [batch-size]
#

set -e

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Batch size (padrão: 100, pode ser passado como argumento)
BATCH_SIZE="${1:-100}"

print_header() {
    echo ""
    echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
    echo ""
}

print_step() {
    echo -e "${GREEN}▶${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Verificar se as variáveis de ambiente estão definidas
check_env() {
    print_step "Verificando variáveis de ambiente..."
    
    if [ -z "$FLOWS_CODE_ACTIONS_MONGO_DB_URI" ]; then
        print_warning "FLOWS_CODE_ACTIONS_MONGO_DB_URI não definida, usando padrão: mongodb://localhost:27017"
        export FLOWS_CODE_ACTIONS_MONGO_DB_URI="mongodb://localhost:27017"
    fi
    
    if [ -z "$FLOWS_CODE_ACTIONS_MONGO_DB_NAME" ]; then
        print_warning "FLOWS_CODE_ACTIONS_MONGO_DB_NAME não definida, usando padrão: code-actions"
        export FLOWS_CODE_ACTIONS_MONGO_DB_NAME="code-actions"
    fi
    
    if [ -z "$FLOWS_CODE_ACTIONS_DB_URI" ]; then
        print_warning "FLOWS_CODE_ACTIONS_DB_URI não definida, usando padrão: postgres://test:test@localhost:5432/codeactions?sslmode=disable"
        export FLOWS_CODE_ACTIONS_DB_URI="postgres://test:test@localhost:5432/codeactions?sslmode=disable"
    fi
    
    print_info "MongoDB: ${FLOWS_CODE_ACTIONS_MONGO_DB_URI}"
    print_info "Database: ${FLOWS_CODE_ACTIONS_MONGO_DB_NAME}"
    print_info "PostgreSQL: $(echo $FLOWS_CODE_ACTIONS_DB_URI | sed 's/:\/\/[^:]*:[^@]*@/:\/\/***:***@/')"
    print_info "Batch size: ${BATCH_SIZE}"
    echo ""
}

# Verificar se os scripts existem
check_scripts() {
    print_step "Verificando scripts..."
    
    if [ ! -f "scripts/export-mongo-to-csv.sh" ]; then
        print_error "Script export-mongo-to-csv.sh não encontrado!"
        exit 1
    fi
    
    if [ ! -f "scripts/import-csv-to-postgres.go" ]; then
        print_error "Script import-csv-to-postgres.go não encontrado!"
        exit 1
    fi
    
    print_success "Scripts encontrados"
    echo ""
}

# Verificar conectividade
check_connectivity() {
    print_step "Verificando conectividade..."
    
    # Testar MongoDB
    print_info "Testando MongoDB..."
    if command -v mongosh &> /dev/null; then
        if mongosh "$FLOWS_CODE_ACTIONS_MONGO_DB_URI" --eval "db.adminCommand('ping')" --quiet &> /dev/null; then
            print_success "MongoDB acessível"
        else
            print_error "Não foi possível conectar ao MongoDB"
            exit 1
        fi
    else
        print_warning "mongosh não disponível, pulando teste de conexão MongoDB"
    fi
    
    # Testar PostgreSQL
    print_info "Testando PostgreSQL..."
    if command -v psql &> /dev/null; then
        if psql "$FLOWS_CODE_ACTIONS_DB_URI" -c "SELECT 1" &> /dev/null; then
            print_success "PostgreSQL acessível"
        else
            print_error "Não foi possível conectar ao PostgreSQL"
            exit 1
        fi
    else
        print_warning "psql não disponível, pulando teste de conexão PostgreSQL"
    fi
    
    echo ""
}

# Mostrar estatísticas pré-migração
show_pre_migration_stats() {
    print_step "Estatísticas pré-migração..."
    
    # MongoDB
    echo -e "${BLUE}MongoDB:${NC}"
    if command -v mongosh &> /dev/null; then
        mongosh "$FLOWS_CODE_ACTIONS_MONGO_DB_URI" --eval "
            db = db.getSiblingDB('$FLOWS_CODE_ACTIONS_MONGO_DB_NAME');
            print('  code: ' + db.code.countDocuments());
            print('  codelib: ' + db.codelib.countDocuments());
            print('  coderun: ' + db.coderun.countDocuments());
        " --quiet 2>/dev/null || print_warning "Não foi possível obter estatísticas do MongoDB"
    else
        print_warning "mongosh não disponível"
    fi
    
    # PostgreSQL
    echo ""
    echo -e "${BLUE}PostgreSQL:${NC}"
    if command -v psql &> /dev/null; then
        psql "$FLOWS_CODE_ACTIONS_DB_URI" -t -c "
            SELECT '  codes: ' || COUNT(*) FROM codes
            UNION ALL
            SELECT '  codelibs: ' || COUNT(*) FROM codelibs
            UNION ALL
            SELECT '  coderuns: ' || COUNT(*) FROM coderuns;
        " 2>/dev/null || print_warning "Não foi possível obter estatísticas do PostgreSQL"
    else
        print_warning "psql não disponível"
    fi
    
    echo ""
}

# Exportar dados do MongoDB
export_data() {
    print_header "ETAPA 1: Exportando dados do MongoDB"
    
    print_step "Executando export-mongo-to-csv.sh..."
    echo ""
    
    if bash scripts/export-mongo-to-csv.sh; then
        echo ""
        print_success "Exportação concluída com sucesso!"
        
        # Encontrar o diretório mais recente
        EXPORT_DIR=$(ls -td mongo_exports/*/ 2>/dev/null | head -1)
        if [ -z "$EXPORT_DIR" ]; then
            print_error "Nenhum diretório de exportação encontrado!"
            exit 1
        fi
        
        # Remover trailing slash
        EXPORT_DIR="${EXPORT_DIR%/}"
        
        print_info "Dados exportados para: $EXPORT_DIR"
        
        # Mostrar tamanho dos arquivos
        echo ""
        echo "Arquivos gerados:"
        ls -lh "$EXPORT_DIR" | tail -n +2 | awk '{print "  " $9 " (" $5 ")"}'
        
        # Contar linhas (registros) em cada arquivo
        echo ""
        echo "Registros por arquivo (incluindo cabeçalho):"
        for file in "$EXPORT_DIR"/*.csv; do
            if [ -f "$file" ]; then
                count=$(wc -l < "$file")
                echo "  $(basename "$file"): $((count - 1)) registros"
            fi
        done
    else
        print_error "Erro na exportação!"
        exit 1
    fi
    
    echo ""
}

# Importar dados para PostgreSQL
import_data() {
    print_header "ETAPA 2: Importando dados para PostgreSQL"
    
    print_step "Executando import-csv-to-postgres.go..."
    print_info "Usando batch-size: $BATCH_SIZE"
    echo ""
    
    if go run scripts/import-csv-to-postgres.go \
        -dir="$EXPORT_DIR" \
        -batch-size="$BATCH_SIZE"; then
        echo ""
        print_success "Importação concluída com sucesso!"
    else
        print_error "Erro na importação!"
        echo ""
        print_warning "Se o erro foi por falta de memória, tente novamente com batch-size menor:"
        echo "  $0 50"
        echo "  ou"
        echo "  $0 10"
        exit 1
    fi
    
    echo ""
}

# Mostrar estatísticas pós-migração
show_post_migration_stats() {
    print_header "ESTATÍSTICAS PÓS-MIGRAÇÃO"
    
    print_step "Comparando registros..."
    echo ""
    
    # PostgreSQL
    echo -e "${BLUE}PostgreSQL (após migração):${NC}"
    if command -v psql &> /dev/null; then
        psql "$FLOWS_CODE_ACTIONS_DB_URI" -t -c "
            SELECT '  codes: ' || COUNT(*) FROM codes
            UNION ALL
            SELECT '  codelibs: ' || COUNT(*) FROM codelibs
            UNION ALL
            SELECT '  coderuns: ' || COUNT(*) FROM coderuns;
        " 2>/dev/null || print_warning "Não foi possível obter estatísticas"
    fi
    
    echo ""
    
    # Verificar integridade
    print_step "Verificando integridade..."
    if command -v psql &> /dev/null; then
        echo ""
        psql "$FLOWS_CODE_ACTIONS_DB_URI" -c "
            SELECT 
                'Códigos sem execuções' as verificacao,
                COUNT(*) as total
            FROM codes c
            WHERE NOT EXISTS (
                SELECT 1 FROM coderuns cr WHERE cr.code_id = c.id
            )
            UNION ALL
            SELECT 
                'Execuções órfãs',
                COUNT(*)
            FROM coderuns cr
            WHERE code_id IS NULL OR NOT EXISTS (
                SELECT 1 FROM codes c WHERE c.id = cr.code_id
            );
        " 2>/dev/null
    fi
    
    echo ""
}

# Limpar arquivos temporários (opcional)
cleanup() {
    echo ""
    echo -e "${YELLOW}Deseja manter os arquivos CSV exportados?${NC}"
    echo "Diretório: $EXPORT_DIR"
    echo ""
    echo "1) Sim, manter (padrão)"
    echo "2) Não, remover"
    echo ""
    read -p "Escolha (1-2): " -t 10 choice || choice=1
    
    case $choice in
        2)
            print_info "Removendo arquivos CSV..."
            rm -rf "$EXPORT_DIR"
            print_success "Arquivos removidos"
            ;;
        *)
            print_info "Arquivos mantidos em: $EXPORT_DIR"
            ;;
    esac
}

# Função principal
main() {
    print_header "Migração MongoDB → PostgreSQL"
    
    echo "Batch size: $BATCH_SIZE"
    echo ""
    echo "Pressione Ctrl+C para cancelar ou Enter para continuar..."
    read -t 5 || true
    echo ""
    
    check_env
    check_scripts
    check_connectivity
    show_pre_migration_stats
    
    export_data
    import_data
    
    show_post_migration_stats
    
    print_header "MIGRAÇÃO CONCLUÍDA!"
    
    print_success "Dados migrados com sucesso do MongoDB para o PostgreSQL!"
    echo ""
    print_info "Próximos passos:"
    echo "  1. Verifique os dados no PostgreSQL"
    echo "  2. Teste a aplicação com o PostgreSQL"
    echo "  3. Ajuste FLOWS_CODE_ACTIONS_DB_TYPE=postgres"
    echo ""
    
    cleanup
}

# Executar
main

