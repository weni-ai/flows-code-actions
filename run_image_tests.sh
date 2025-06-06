#!/bin/bash

# Script para executar testes de integração contra imagem Docker da aplicação
set -e

echo "🐳 Teste de Integração - Imagem Docker"
echo "======================================"

# Configurações
APP_IMAGE_NAME="codeactions-app"
APP_CONTAINER_NAME="codeactions-app-test"
APP_PORT="8050"
BASE_URL="http://localhost:$APP_PORT"
NETWORK_NAME="codeactions-test-network"

# Função para verificar Docker
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker não está instalado"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        echo "❌ Docker não está rodando"
        exit 1
    fi
    
    echo "✅ Docker está funcionando"
}

# Função para limpar recursos
cleanup() {
    echo "🧹 Limpando containers e rede..."
    
    # Parar e remover containers
    echo "🛑 Parando containers..."
    docker stop $APP_CONTAINER_NAME mongo-test redis-test 2>/dev/null || true
    
    echo "🗑️  Removendo containers..."
    docker rm $APP_CONTAINER_NAME mongo-test redis-test 2>/dev/null || true
    
    # Remover rede
    echo "🌐 Removendo rede..."
    docker network rm $NETWORK_NAME 2>/dev/null || true
    
    echo "✅ Limpeza concluída"
}

# Função para limpeza inicial mais robusta
initial_cleanup() {
    echo "🧹 Limpeza inicial..."
    
    # Parar containers que possam estar rodando
    docker stop $APP_CONTAINER_NAME mongo-test redis-test 2>/dev/null || true
    docker rm $APP_CONTAINER_NAME mongo-test redis-test 2>/dev/null || true
    
    # Remover rede anterior
    docker network rm $NETWORK_NAME 2>/dev/null || true
    
    # Limpar containers órfãos relacionados ao projeto
    docker container prune -f --filter "label=project=codeactions-test" 2>/dev/null || true
    
    echo "✅ Limpeza inicial concluída"
}

# Função para construir imagem da aplicação
build_app_image() {
    echo "🔨 Construindo imagem da aplicação..."
    
    if [ "$1" = "--force" ]; then
        echo "🔄 Forçando rebuild da imagem..."
        docker build --no-cache -t $APP_IMAGE_NAME .
    else
        # Verificar se imagem existe
        if docker images | grep -q "$APP_IMAGE_NAME"; then
            echo "✅ Imagem $APP_IMAGE_NAME já existe"
            return 0
        fi
        
        echo "📦 Construindo nova imagem..."
        docker build -t $APP_IMAGE_NAME .
    fi
    
    echo "✅ Imagem $APP_IMAGE_NAME construída"
}

# Função para iniciar serviços de dependência
start_dependencies() {
    echo "📦 Iniciando dependências (MongoDB e Redis)..."
    
    # Criar rede se não existir
    docker network create $NETWORK_NAME 2>/dev/null || true
    
    # Parar e remover containers existentes se existirem
    docker stop mongo-test redis-test 2>/dev/null || true
    docker rm mongo-test redis-test 2>/dev/null || true
    
    # MongoDB
    echo "🗄️  Iniciando MongoDB..."
    docker run -d \
        --name mongo-test \
        --network $NETWORK_NAME \
        -e MONGO_INITDB_DATABASE=codeactions_test \
        mongo:7
    
    # Redis
    echo "🔴 Iniciando Redis..."
    docker run -d \
        --name redis-test \
        --network $NETWORK_NAME \
        redis:7-alpine
    
    echo "⏳ Aguardando dependências ficarem prontas..."
    
    # Aguardar MongoDB
    local retries=30
    echo "🔍 Testando conexão MongoDB..."
    for i in $(seq 1 $retries); do
        if docker exec mongo-test mongosh --eval "db.adminCommand('ping')" --quiet &>/dev/null; then
            echo "✅ MongoDB pronto"
            break
        fi
        if [ $i -eq $retries ]; then
            echo "❌ MongoDB não ficou pronto"
            echo "📋 Logs do MongoDB:"
            docker logs mongo-test --tail 20
            exit 1
        fi
        echo "⏳ MongoDB - tentativa $i/$retries..."
        sleep 3
    done
    
    # Aguardar Redis
    echo "🔍 Testando conexão Redis..."
    for i in $(seq 1 $retries); do
        if docker exec redis-test redis-cli ping | grep -q "PONG"; then
            echo "✅ Redis pronto"
            break
        fi
        if [ $i -eq $retries ]; then
            echo "❌ Redis não ficou pronto"
            echo "📋 Logs do Redis:"
            docker logs redis-test --tail 20
            exit 1
        fi
        echo "⏳ Redis - tentativa $i/$retries..."
        sleep 2
    done
    
    # Verificar se containers estão na rede correta
    echo "🔍 Verificando conectividade na rede..."
    docker network inspect $NETWORK_NAME --format '{{range .Containers}}{{.Name}} {{end}}' | grep -q "mongo-test\|redis-test"
    if [ $? -eq 0 ]; then
        echo "✅ Containers estão na rede $NETWORK_NAME"
    else
        echo "❌ Problema na configuração da rede"
        docker network inspect $NETWORK_NAME
        exit 1
    fi
}

# Função para iniciar aplicação
start_application() {
    echo "🚀 Iniciando aplicação..."
    
    # Parar e remover container da aplicação se existir
    docker stop $APP_CONTAINER_NAME 2>/dev/null || true
    docker rm $APP_CONTAINER_NAME 2>/dev/null || true
    
    # Variáveis de ambiente para aplicação
    echo "🔧 Configurando variáveis de ambiente:"
    echo "   - MongoDB URI: mongodb://mongo-test:27017"
    echo "   - MongoDB Name: codeactions_test"
    echo "   - Redis: redis://redis-test:6379/1"
    echo "   - Port: $APP_PORT"
    
    docker run -d \
        --name $APP_CONTAINER_NAME \
        --network $NETWORK_NAME \
        -p $APP_PORT:$APP_PORT \
        -e FLOWS_CODE_ACTIONS_MONGO_DB_URI="mongodb://mongo-test:27017" \
        -e FLOWS_CODE_ACTIONS_MONGO_DB_NAME="codeactions_test" \
        -e FLOWS_CODE_ACTIONS_REDIS="redis://redis-test:6379/1" \
        -e FLOWS_CODE_ACTIONS_HTTP_PORT="$APP_PORT" \
        -e FLOWS_CODE_ACTIONS_ENVIRONMENT="test" \
        $APP_IMAGE_NAME
    
    echo "⏳ Aguardando aplicação ficar disponível..."
    
    # Aguardar aplicação responder
    local retries=60  # Aumentar timeout para debug
    for i in $(seq 1 $retries); do
        # Verificar se container ainda está rodando
        if ! docker ps --format "table {{.Names}}" | grep -q "$APP_CONTAINER_NAME"; then
            echo "❌ Container da aplicação parou de funcionar!"
            echo "📋 Logs da aplicação:"
            docker logs $APP_CONTAINER_NAME --tail 50
            exit 1
        fi
        
        # Testar health endpoint
        if curl -s -f "$BASE_URL/health" &>/dev/null; then
            echo "✅ Aplicação disponível em $BASE_URL"
            return 0
        fi
        
        if [ $i -eq $retries ]; then
            echo "❌ Aplicação não ficou disponível"
            echo "📋 Status do container:"
            docker ps -a --filter name=$APP_CONTAINER_NAME
            echo ""
            echo "📋 Logs da aplicação:"
            docker logs $APP_CONTAINER_NAME --tail 50
            echo ""
            echo "📋 Teste de conectividade na rede:"
            docker exec $APP_CONTAINER_NAME ping -c 2 mongo-test || true
            docker exec $APP_CONTAINER_NAME ping -c 2 redis-test || true
            exit 1
        fi
        
        if [ $((i % 10)) -eq 0 ]; then
            echo "⏳ Tentativa $i/$retries - Aguardando aplicação..."
            echo "📋 Status atual do container:"
            docker ps --filter name=$APP_CONTAINER_NAME --format "table {{.Names}}\t{{.Status}}"
        fi
        
        sleep 2
    done
}

# Função para executar testes
run_tests() {
    echo "🧪 Executando testes de integração..."
    
    # Configurar variável de ambiente para os testes
    export CODEACTIONS_BASE_URL="$BASE_URL"
    
    local test_cmd="go test -v -timeout 15m ./integration_image_test.go"
    
    case "$1" in
        "verbose")
            echo "🔍 Modo verboso ativado"
            $test_cmd
            ;;
        "specific")
            echo "🎯 Executando teste específico: $2"
            go test -v -timeout 15m -run "$2" ./integration_image_test.go
            ;;
        *)
            echo "🏃 Executando todos os testes..."
            $test_cmd
            ;;
    esac
    
    local test_exit_code=$?
    
    if [ $test_exit_code -eq 0 ]; then
        echo "🎉 Todos os testes passaram!"
    else
        echo "❌ Alguns testes falharam!"
        echo ""
        echo "📋 Logs da aplicação:"
        docker logs $APP_CONTAINER_NAME --tail 50
    fi
    
    return $test_exit_code
}

# Função para mostrar logs
show_logs() {
    echo "📋 Logs dos serviços:"
    echo "===================="
    
    echo "🔸 Aplicação:"
    docker logs $APP_CONTAINER_NAME --tail 20
    
    echo ""
    echo "🔸 MongoDB:"
    docker logs mongo-test --tail 10
    
    echo ""
    echo "🔸 Redis:"
    docker logs redis-test --tail 10
}

# Função principal
main() {
    local force_build=false
    local test_mode="normal"
    local specific_test=""
    local cleanup_after=true
    local show_logs_flag=false
    
    # Processar argumentos
    while [[ $# -gt 0 ]]; do
        case $1 in
            --build|-b)
                force_build=true
                shift
                ;;
            --verbose|-v)
                test_mode="verbose"
                shift
                ;;
            --test|-t)
                test_mode="specific"
                specific_test="$2"
                shift 2
                ;;
            --no-cleanup)
                cleanup_after=false
                shift
                ;;
            --logs|-l)
                show_logs_flag=true
                shift
                ;;
            --port|-p)
                APP_PORT="$2"
                BASE_URL="http://localhost:$APP_PORT"
                shift 2
                ;;
            --help|-h)
                echo "Uso: $0 [opções]"
                echo ""
                echo "Opções:"
                echo "  --build, -b         Forçar rebuild da imagem"
                echo "  --verbose, -v       Executar testes em modo verboso"
                echo "  --test, -t <nome>   Executar teste específico"
                echo "  --no-cleanup        Não limpar containers após testes"
                echo "  --logs, -l          Mostrar logs dos serviços"
                echo "  --port, -p <porta>  Porta da aplicação (padrão: 8050)"
                echo "  --debug-crash       Investigar crashes da aplicação"
                echo "  --help, -h          Mostrar esta ajuda"
                echo ""
                echo "Exemplos:"
                echo "  $0                                          # Executar todos os testes"
                echo "  $0 --build                                 # Rebuild e executar"
                echo "  $0 --verbose                               # Modo verboso"
                echo "  $0 --test TestCreateAndExecuteCode        # Teste específico"
                echo "  $0 --no-cleanup --logs                    # Manter containers e mostrar logs"
                echo "  $0 --port 9000                            # Usar porta personalizada"
                echo "  $0 --debug-crash                          # Debug de crashes"
                exit 0
                ;;
            --debug-crash)
                echo "🔍 Modo debug de crash ativado"
                echo "Iniciando containers e investigando problemas..."
                
                # Verificar dependências
                check_docker
                
                # Limpar ambiente
                initial_cleanup
                
                # Construir imagem se necessário
                build_app_image
                
                # Iniciar dependências
                start_dependencies
                
                # Iniciar aplicação
                start_application
                
                # Executar debug específico
                ./debug_application_crash.sh
                
                echo ""
                echo "💡 Para continuar debugando:"
                echo "   docker exec -it $APP_CONTAINER_NAME sh"
                echo "   docker logs $APP_CONTAINER_NAME -f"
                
                exit 0
                ;;
            *)
                echo "❌ Opção desconhecida: $1"
                echo "Use --help para ver as opções disponíveis"
                exit 1
                ;;
        esac
    done
    
    # Verificar dependências
    check_docker
    
    if ! command -v curl &> /dev/null; then
        echo "⚠️  curl não está instalado (recomendado para verificação de saúde)"
    fi
    
    # Configurar trap para cleanup
    if [ "$cleanup_after" = true ]; then
        trap cleanup EXIT
    fi
    
    # Limpar ambiente anterior
    initial_cleanup
    
    # Construir imagem
    if [ "$force_build" = true ]; then
        build_app_image --force
    else
        build_app_image
    fi
    
    # Iniciar dependências
    start_dependencies
    
    # Iniciar aplicação
    start_application
    
    # Executar testes
    if [ "$test_mode" = "specific" ]; then
        run_tests "$test_mode" "$specific_test"
    else
        run_tests "$test_mode"
    fi
    
    local test_result=$?
    
    # Mostrar logs se solicitado ou se houve falha
    if [ "$show_logs_flag" = true ] || [ $test_result -ne 0 ]; then
        show_logs
    fi
    
    return $test_result
}

# Executar função principal
main "$@" 