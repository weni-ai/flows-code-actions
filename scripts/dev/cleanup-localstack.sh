#!/bin/bash

# Script para limpar completamente LocalStack e resolver problemas de "Device or resource busy"

set -e

echo "🧹 Limpando LocalStack completamente..."

# Para todos os containers LocalStack
echo "🛑 Parando containers LocalStack..."
docker ps -a --filter "name=localstack" --format "table {{.Names}}" | tail -n +2 | xargs -r docker stop 2>/dev/null || true
docker ps -a --filter "ancestor=localstack/localstack" --format "table {{.Names}}" | tail -n +2 | xargs -r docker stop 2>/dev/null || true

# Remove containers LocalStack
echo "🗑️  Removendo containers LocalStack..."
docker ps -a --filter "name=localstack" --format "table {{.Names}}" | tail -n +2 | xargs -r docker rm -f 2>/dev/null || true
docker ps -a --filter "ancestor=localstack/localstack" --format "table {{.Names}}" | tail -n +2 | xargs -r docker rm -f 2>/dev/null || true

# Para docker-compose específico
echo "🏗️  Parando docker-compose..."
docker compose -f docker-compose.localstack.yml down -v 2>/dev/null || true

# Remove volumes relacionados ao LocalStack
echo "📁 Removendo volumes LocalStack..."
docker volume ls -q | grep -i localstack | xargs -r docker volume rm 2>/dev/null || true

# Limpa volumes órfãos
echo "🧽 Limpando volumes órfãos..."
docker volume prune -f 2>/dev/null || true

# Remove redes órfãs
echo "🌐 Limpando redes órfãs..."
docker network prune -f 2>/dev/null || true

# Força limpeza do diretório /tmp/localstack se existir
echo "🗂️  Limpando diretório temporário..."
if [ -d "/tmp/localstack" ]; then
    echo "   Tentando remover /tmp/localstack..."
    # Primeiro tenta umount se necessário
    sudo umount /tmp/localstack 2>/dev/null || true
    # Tenta remover normalmente
    sudo rm -rf /tmp/localstack 2>/dev/null || {
        echo "   ⚠️  Não foi possível remover /tmp/localstack diretamente"
        echo "   💡 Isso é normal, será criado um novo diretório"
    }
fi

echo ""
echo "✅ Limpeza concluída!"
echo ""
echo "📋 Próximos passos:"
echo "   1. Execute: ./scripts/start-localstack.sh"
echo "   2. Se ainda der erro, reinicie o Docker:"
echo "      sudo systemctl restart docker"
echo "   3. Ou reinicie o computador se necessário"
