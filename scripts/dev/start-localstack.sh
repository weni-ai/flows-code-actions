#!/bin/bash

# Script para iniciar LocalStack e configurar a aplicação para desenvolvimento S3

set -e

echo "🚀 Iniciando LocalStack para desenvolvimento S3..."

# Verifica se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Por favor, inicie o Docker primeiro."
    exit 1
fi

# Limpa containers existentes completamente
echo "🛑 Parando containers LocalStack existentes..."
docker compose -f docker-compose.localstack.yml down -v 2>/dev/null || true

# Remove containers LocalStack antigos se houver
echo "🧹 Limpando containers antigos..."
docker ps -a --filter "name=localstack" -q | xargs -r docker rm -f 2>/dev/null || true

# Inicia LocalStack
echo "🏗️  Iniciando LocalStack..."
docker compose -f docker-compose.localstack.yml up -d

# Aguarda LocalStack estar pronto
echo "⏳ Aguardando LocalStack estar pronto..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:4566/_localstack/health | grep -q '"s3": "available"'; then
        echo "✅ LocalStack está pronto!"
        break
    fi
    
    attempt=$((attempt + 1))
    echo "   Tentativa $attempt/$max_attempts - aguardando..."
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ LocalStack não ficou pronto no tempo esperado"
    exit 1
fi

# Executa script de inicialização
echo "🗂️  Configurando bucket S3..."
bash scripts/dev/init-s3.sh

echo ""
echo "🎉 LocalStack configurado com sucesso!"
echo ""
echo "📋 Próximos passos:"
echo "   1. Configure as variáveis de ambiente:"
echo "      source config.localstack.example"
echo ""
echo "   2. Inicie sua aplicação normalmente"
echo ""
echo "🔗 URLs úteis:"
echo "   - LocalStack health: http://localhost:4566/health"
echo "   - S3 endpoint: http://localhost:4566"
echo ""
echo "🧪 Testando configuração:"
echo "   aws --endpoint-url=http://localhost:4566 s3 ls s3://codeactions-dev"
echo ""
echo "💡 Se tiver problemas com 'Device or resource busy':"
echo "   ./scripts/cleanup-localstack.sh"

