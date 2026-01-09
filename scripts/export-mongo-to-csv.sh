#!/bin/bash
#
# Script para exportar dados do MongoDB para CSV
# Este script exporta todas as collections relevantes do MongoDB para arquivos CSV
#
# Uso: ./export-mongo-to-csv.sh
#

set -e

# Configurações do MongoDB
MONGO_URI="${FLOWS_CODE_ACTIONS_MONGO_DB_URI:-mongodb://localhost:27017}"
MONGO_DB="${FLOWS_CODE_ACTIONS_MONGO_DB_NAME:-code-actions}"

# Diretório de saída
OUTPUT_DIR="./mongo_exports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
EXPORT_DIR="${OUTPUT_DIR}/${TIMESTAMP}"

# Criar diretório de saída
mkdir -p "${EXPORT_DIR}"

echo "============================================"
echo "Exportando dados do MongoDB para CSV"
echo "============================================"
echo "MongoDB URI: ${MONGO_URI}"
echo "Database: ${MONGO_DB}"
echo "Diretório de saída: ${EXPORT_DIR}"
echo "============================================"
echo ""

# Função para exportar uma collection
export_collection() {
    local collection=$1
    local fields=$2
    local output_file="${EXPORT_DIR}/${collection}.csv"
    
    echo "Exportando collection: ${collection}..."
    
    mongoexport \
        --uri="${MONGO_URI}/${MONGO_DB}" \
        --collection="${collection}" \
        --type=csv \
        --fields="${fields}" \
        --out="${output_file}"
    
    if [ $? -eq 0 ]; then
        local count=$(wc -l < "${output_file}")
        count=$((count - 1)) # Subtrair linha de cabeçalho
        echo "✓ ${collection} exportado com sucesso (${count} registros)"
        echo ""
    else
        echo "✗ Erro ao exportar ${collection}"
        echo ""
    fi
}

# Exportar collection: code
# Campos: _id, name, type, source, language, url, project_uuid, timeout, created_at, updated_at
export_collection "code" "_id,name,type,source,language,url,project_uuid,timeout,created_at,updated_at"

# Exportar collection: codelib
# Campos: _id, name, language, created_at, updated_at
export_collection "codelib" "_id,name,language,created_at,updated_at"

# Exportar collection: coderun
# Campos: _id, code_id, status, result, extra, params, body, headers, created_at, updated_at
export_collection "coderun" "_id,code_id,status,result,extra,params,body,headers,created_at,updated_at"

# Exportar collection: user_permissions (se existir)
# Campos: _id, user_email, project_uuid, role, created_at, updated_at
echo "Tentando exportar user_permissions..."
mongoexport \
    --uri="${MONGO_URI}/${MONGO_DB}" \
    --collection="user_permissions" \
    --type=csv \
    --fields="_id,user_email,project_uuid,role,created_at,updated_at" \
    --out="${EXPORT_DIR}/user_permissions.csv" 2>/dev/null || echo "Collection user_permissions não encontrada ou vazia"

echo ""

# Exportar collection: projects (se existir)
# Campos: _id, uuid, name, created_at, updated_at
echo "Tentando exportar projects..."
mongoexport \
    --uri="${MONGO_URI}/${MONGO_DB}" \
    --collection="projects" \
    --type=csv \
    --fields="_id,uuid,name,created_at,updated_at" \
    --out="${EXPORT_DIR}/projects.csv" 2>/dev/null || echo "Collection projects não encontrada ou vazia"

echo ""
echo "============================================"
echo "Exportação concluída!"
echo "============================================"
echo "Arquivos gerados em: ${EXPORT_DIR}"
echo ""
echo "Arquivos exportados:"
ls -lh "${EXPORT_DIR}"
echo ""
echo "Para importar no PostgreSQL, execute:"
echo "  go run scripts/import-csv-to-postgres.go -dir=${EXPORT_DIR}"
echo ""

