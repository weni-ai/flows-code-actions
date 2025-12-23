#!/bin/bash

# Script alternativo para iniciar LocalStack com proteção contra "Device or resource busy"

set -e

echo "🚀 Iniciando LocalStack (modo seguro)..."

# Verifica se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Por favor, inicie o Docker primeiro."
    exit 1
fi

# Função para limpar LocalStack
cleanup_localstack() {
    echo "🧹 Limpando LocalStack..."
    
    # Para e remove containers
    docker ps -a --filter "name=localstack" -q | xargs -r docker rm -f 2>/dev/null || true
    docker ps -a --filter "ancestor=localstack/localstack" -q | xargs -r docker rm -f 2>/dev/null || true
    
    # Remove volumes órfãos
    docker volume prune -f 2>/dev/null || true
    
    echo "✅ Limpeza concluída"
}

# Tenta iniciar LocalStack
start_localstack() {
    echo "🏗️  Iniciando LocalStack..."
    
    # Inicia com docker run para mais controle
    docker run -d \
        --name localstack-s3 \
        --rm \
        -p 4566:4566 \
        -e SERVICES=s3 \
        -e S3_SKIP_SIGNATURE_VALIDATION=1 \
        -e S3_SKIP_KMS_KEY_VALIDATION=1 \
        -e DEBUG=0 \
        -v "$PWD/localstack:/docker-entrypoint-initaws.d" \
        localstack/localstack:3.0
        
    return $?
}

# Limpa primeiro
cleanup_localstack

# Tenta iniciar
if ! start_localstack; then
    echo "❌ Falha ao iniciar LocalStack com docker run"
    echo "🔄 Tentando com docker-compose..."
    
    # Fallback para docker-compose
    docker compose -f docker-compose.localstack.yml up -d
fi

# Aguarda LocalStack estar pronto
echo "⏳ Aguardando LocalStack estar pronto..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:4566/_localstack/health | grep -q '"s3": "available"' 2>/dev/null; then
        echo "✅ LocalStack está pronto!"
        break
    fi
    
    attempt=$((attempt + 1))
    echo "   Tentativa $attempt/$max_attempts - aguardando..."
    sleep 2
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ LocalStack não ficou pronto no tempo esperado"
    echo "🔍 Verificando logs..."
    docker logs localstack-s3 2>/dev/null || docker compose -f docker-compose.localstack.yml logs
    exit 1
fi

# Executa script de inicialização
echo "🗂️  Configurando bucket S3..."
bash localstack/init-s3.sh

echo ""
echo "🎉 LocalStack configurado com sucesso!"
echo ""
echo "📋 Para usar:"
echo "   source config.localstack.example"
echo "   air -d"
echo ""
echo "🔗 URLs:"
echo "   - Health: http://localhost:4566/health"
echo "   - S3: http://localhost:4566"
