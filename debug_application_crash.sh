#!/bin/bash

# Script de debug para investigar crashes da aplicação
echo "🔍 Debug - Crash da Aplicação"
echo "============================="

NETWORK_NAME="codeactions-test-network"
APP_CONTAINER_NAME="codeactions-app-test"

echo "📋 Verificando status atual dos containers:"
echo "==========================================="
docker ps -a --filter name=mongo-test
docker ps -a --filter name=redis-test
docker ps -a --filter name=$APP_CONTAINER_NAME

echo ""
echo "📜 Logs completos da aplicação (últimas 100 linhas):"
echo "=================================================="
if docker ps -a --filter name=$APP_CONTAINER_NAME | grep -q $APP_CONTAINER_NAME; then
    docker logs $APP_CONTAINER_NAME --tail 100
else
    echo "❌ Container da aplicação não encontrado"
    exit 1
fi

echo ""
echo "🔍 Investigando linha 167 de service.go:"
echo "========================================"
echo "A linha 167 está no método runPython, especificamente em:"
echo "  → cmd.Wait() - esperando processo Python terminar"
echo "  → Isso indica que o Python está falhando ou crashando"

echo ""
echo "🐍 Verificando ambiente Python no container:"
echo "==========================================="
if docker ps --filter name=$APP_CONTAINER_NAME | grep -q $APP_CONTAINER_NAME; then
    echo "📋 Versão do Python:"
    docker exec $APP_CONTAINER_NAME python --version || echo "❌ Python não encontrado"
    
    echo ""
    echo "📋 Verificando engines/py/main.py:"
    docker exec $APP_CONTAINER_NAME ls -la /app/engines/py/ || echo "❌ Diretório engines/py não encontrado"
    
    echo ""
    echo "📋 Testando Python básico:"
    docker exec $APP_CONTAINER_NAME python -c "print('Python OK')" || echo "❌ Python não funciona"
    
    echo ""
    echo "📋 Verificando bibliotecas Python necessárias:"
    docker exec $APP_CONTAINER_NAME python -c "import json, sys, argparse, pymongo; print('Bibliotecas básicas OK')" || echo "❌ Bibliotecas básicas falharam"
    
    echo ""
    echo "📋 Espaço em disco no container:"
    docker exec $APP_CONTAINER_NAME df -h
    
    echo ""
    echo "📋 Verificando permissões do diretório de trabalho:"
    docker exec $APP_CONTAINER_NAME ls -la /app/
    docker exec $APP_CONTAINER_NAME whoami
    
    echo ""
    echo "📋 Variáveis de ambiente do container:"
    docker exec $APP_CONTAINER_NAME env | grep -E "(PATH|PYTHON|FLOWS_CODE_ACTIONS)"
else
    echo "❌ Container não está rodando, não é possível investigar"
fi

echo ""
echo "🧪 Testando execução Python manual:"
echo "=================================="
if docker ps --filter name=$APP_CONTAINER_NAME | grep -q $APP_CONTAINER_NAME; then
    echo "Criando teste Python simples..."
    
    # Criar um teste Python simples
    docker exec $APP_CONTAINER_NAME sh -c 'echo "def Run(engine): print(\"Hello from Python\")" > /tmp/test_action.py'
    
    # Tentar executar com o mesmo método da aplicação
    echo "Testando execução Python:"
    docker exec $APP_CONTAINER_NAME python -c "
import sys
sys.path.append('/app/engines/py')
try:
    import main
    print('main.py importado com sucesso')
except Exception as e:
    print(f'Erro ao importar main.py: {e}')
"
fi

echo ""
echo "🔍 Verificando cgroups (Resource Management):"
echo "============================================"
if docker ps --filter name=$APP_CONTAINER_NAME | grep -q $APP_CONTAINER_NAME; then
    echo "📋 Sistema de cgroups:"
    docker exec $APP_CONTAINER_NAME ls -la /sys/fs/cgroup/ || echo "❌ cgroups não disponível"
    
    echo ""
    echo "📋 Verificando se cgroupfs está montado:"
    docker exec $APP_CONTAINER_NAME cat /proc/mounts | grep cgroup || echo "❌ cgroup não montado"
fi

echo ""
echo "🔄 Executando aplicação com logs verbosos:"
echo "=========================================="
echo "Para debug mais detalhado, você pode:"
echo ""
echo "1. Rodar aplicação com logs de debug:"
echo "   docker exec -it $APP_CONTAINER_NAME sh"
echo "   # Dentro do container:"
echo "   export FLOWS_CODE_ACTIONS_LOG_LEVEL=debug"
echo "   /app/codeactions"
echo ""
echo "2. Monitorar logs em tempo real:"
echo "   docker logs $APP_CONTAINER_NAME -f"
echo ""
echo "3. Testar código Python diretamente:"
echo "   docker exec -it $APP_CONTAINER_NAME python /app/engines/py/main.py --help"

echo ""
echo "📊 Recursos do sistema:"
echo "====================="
echo "📋 Memória e CPU do container:"
docker stats $APP_CONTAINER_NAME --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" 2>/dev/null || echo "❌ Stats não disponível"

echo ""
echo "📋 Logs do sistema (dmesg) para verificar OOM kills:"
docker exec $APP_CONTAINER_NAME dmesg | tail -20 2>/dev/null || echo "❌ dmesg não disponível (normal em containers)"

echo ""
echo "💡 Próximos passos recomendados:"
echo "==============================="
echo "1. Se Python não está funcionando:"
echo "   → Verificar se imagem Docker tem Python instalado"
echo "   → Verificar se engines/py/main.py existe"
echo ""
echo "2. Se há problemas de recursos:"
echo "   → Desabilitar Resource Management temporariamente"
echo "   → Aumentar limites de CPU/memória"
echo ""
echo "3. Se há problemas de permissões:"
echo "   → Verificar usuário que roda a aplicação"
echo "   → Verificar permissões dos arquivos"
echo ""
echo "4. Para debug detalhado:"
echo "   → Adicionar mais logs no código Go"
echo "   → Executar aplicação em modo debug"
echo "   → Testar código Python isoladamente" 