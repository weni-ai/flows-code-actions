#!/bin/bash

# Script de debug para diagnosticar problemas com containers
echo "🔍 Debug de Containers - Code Actions"
echo "====================================="

NETWORK_NAME="codeactions-test-network"
APP_CONTAINER_NAME="codeactions-app-test"

echo "📋 Status dos containers:"
echo "========================="
docker ps -a --filter name=mongo-test
docker ps -a --filter name=redis-test
docker ps -a --filter name=$APP_CONTAINER_NAME

echo ""
echo "🌐 Status da rede:"
echo "=================="
docker network ls | grep $NETWORK_NAME || echo "❌ Rede $NETWORK_NAME não encontrada"

if docker network ls | grep -q $NETWORK_NAME; then
    echo "📋 Containers na rede:"
    docker network inspect $NETWORK_NAME --format '{{range .Containers}}{{.Name}} ({{.IPv4Address}}) {{end}}'
fi

echo ""
echo "📋 Logs dos containers:"
echo "======================"

if docker ps -a --filter name=mongo-test | grep -q mongo-test; then
    echo "🗄️  MongoDB (últimas 10 linhas):"
    docker logs mongo-test --tail 10
else
    echo "❌ Container mongo-test não encontrado"
fi

echo ""
if docker ps -a --filter name=redis-test | grep -q redis-test; then
    echo "🔴 Redis (últimas 10 linhas):"
    docker logs redis-test --tail 10
else
    echo "❌ Container redis-test não encontrado"
fi

echo ""
if docker ps -a --filter name=$APP_CONTAINER_NAME | grep -q $APP_CONTAINER_NAME; then
    echo "🚀 Aplicação (últimas 20 linhas):"
    docker logs $APP_CONTAINER_NAME --tail 20
else
    echo "❌ Container $APP_CONTAINER_NAME não encontrado"
fi

echo ""
echo "🔍 Testes de conectividade:"
echo "=========================="

# Testar conexão a partir do container da aplicação (se existir)
if docker ps --filter name=$APP_CONTAINER_NAME | grep -q $APP_CONTAINER_NAME; then
    echo "🔧 Variáveis de ambiente da aplicação:"
    docker exec $APP_CONTAINER_NAME env | grep FLOWS_CODE_ACTIONS || echo "❌ Variáveis FLOWS_CODE_ACTIONS não encontradas"
    
    echo ""
    echo "📡 Testando ping para MongoDB..."
    docker exec $APP_CONTAINER_NAME ping -c 2 mongo-test || echo "❌ Falha no ping para mongo-test"
    
    echo "📡 Testando ping para Redis..."
    docker exec $APP_CONTAINER_NAME ping -c 2 redis-test || echo "❌ Falha no ping para redis-test"
    
    echo "📡 Testando resolução DNS..."
    docker exec $APP_CONTAINER_NAME nslookup mongo-test || echo "❌ Falha na resolução de mongo-test"
    docker exec $APP_CONTAINER_NAME nslookup redis-test || echo "❌ Falha na resolução de redis-test"
else
    echo "⚠️  Container da aplicação não está rodando, não é possível testar conectividade"
fi

echo ""
echo "🌐 Teste de endpoint externo:"
echo "============================="
curl -s -f "http://localhost:8050/health" && echo "✅ Health endpoint OK" || echo "❌ Health endpoint falhou"

echo ""
echo "💡 Comandos úteis para debug:"
echo "============================"
echo "# Ver logs em tempo real:"
echo "docker logs $APP_CONTAINER_NAME -f"
echo ""
echo "# Entrar no container da aplicação:"
echo "docker exec -it $APP_CONTAINER_NAME sh"
echo ""
echo "# Testar conectividade manual:"
echo "docker exec $APP_CONTAINER_NAME ping mongo-test"
echo "docker exec $APP_CONTAINER_NAME telnet mongo-test 27017"
echo ""
echo "# Limpar tudo:"
echo "./run_image_tests.sh --help" 