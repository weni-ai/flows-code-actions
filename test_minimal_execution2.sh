#!/bin/bash

# Teste mínimo para verificar execução de código Python com logs
echo "🧪 Teste Mínimo - Código com Logs"
echo "================================="

BASE_URL="http://localhost:8050"
PROJECT_UUID="test-logs-$(date +%s)"
CLEANUP_FLAG=false

# Verificar flags
while [[ $# -gt 0 ]]; do
  case $1 in
    -d|--delete)
      CLEANUP_FLAG=true
      shift
      ;;
    *)
      echo "Uso: $0 [-d|--delete]"
      echo "  -d, --delete    Deletar código criado após o teste"
      exit 1
      ;;
  esac
done

echo "🔍 Verificando se aplicação está respondendo..."
if ! curl -s -f "$BASE_URL/health" &>/dev/null; then
    echo "❌ Aplicação não está respondendo em $BASE_URL"
    echo "   Execute: ./run_image_tests.sh --debug-crash"
    exit 1
fi

echo "✅ Aplicação está respondendo"

echo ""
echo "🐍 Testando código Python com geração de logs..."

# Código Python que gera diferentes tipos de logs
LOG_CODE='import json
import datetime

def Run(engine):
    # Gerar diferentes tipos de logs
    engine.log.debug("Debug log: Iniciando processamento")
    engine.log.info("Info log: Carregando dados de entrada")
    
    try:
        # Simulando processamento
        timestamp = datetime.datetime.now().isoformat()
        data = {
            "timestamp": timestamp,
            "calculation": 42 + 8,
            "message": "Processamento concluído com sucesso",
            "status": "success"
        }
        
        engine.log.info(f"Info log: Cálculo realizado: {data[\"calculation\"]}")
        engine.log.debug("Debug log: Preparando resposta")
        
        engine.result.set(data, content_type="json")
        
    except Exception as e:
        engine.log.error(f"Error log: Falha no processamento - {str(e)}")
        raise e'

echo "📝 Código de teste (com logs):"
echo "$LOG_CODE"

echo ""
echo "📤 Criando código via API..."

# Criar código
RESPONSE=$(curl -s -X POST "$BASE_URL/code" \
  -H "Content-Type: text/plain" \
  -d "$LOG_CODE" \
  --get \
  --data-urlencode "project_uuid=$PROJECT_UUID" \
  --data-urlencode "name=Test Code with Logs" \
  --data-urlencode "type=endpoint" \
  --data-urlencode "language=python")

echo "📋 Resposta da criação:"
echo "$RESPONSE"

# Extrair ID do código
CODE_ID=$(echo "$RESPONSE" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(data.get('id', ''))
except:
    pass
" 2>/dev/null)

if [ -z "$CODE_ID" ]; then
    echo "❌ Falha ao criar código ou extrair ID"
    echo "   Resposta completa: $RESPONSE"
    exit 1
fi

echo "✅ Código criado com ID: $CODE_ID"

echo ""
echo "🚀 Executando código para gerar logs..."

# Executar código
EXEC_RESPONSE=$(curl -s -X POST "$BASE_URL/action/endpoint/$CODE_ID" \
  -H "Content-Type: application/json" \
  -w "\nHTTP_STATUS_CODE:%{http_code}")

HTTP_STATUS=$(echo "$EXEC_RESPONSE" | tail -1 | cut -d: -f2)
EXEC_BODY=$(echo "$EXEC_RESPONSE" | head -n -1)

echo "📋 Status HTTP: $HTTP_STATUS"
echo "📋 Resposta da execução:"
echo "$EXEC_BODY"

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Execução bem-sucedida!"
    
    # Verificar se a resposta contém o resultado esperado
    if echo "$EXEC_BODY" | grep -q "success"; then
        echo "✅ Resultado contém dados esperados"
    else
        echo "⚠️  Resultado não contém dados esperados"
    fi
else
    echo "❌ Execução falhou com status $HTTP_STATUS"
fi

echo ""
echo "📜 Aguardando logs serem persistidos..."
sleep 3

echo "📋 Listando logs do código executado..."

# Listar logs por code_id
LOGS_RESPONSE=$(curl -s -G "$BASE_URL/codelog" \
  --data-urlencode "code_id=$CODE_ID" \
  --data-urlencode "page=1" \
  -w "\nHTTP_STATUS_CODE:%{http_code}")

LOGS_HTTP_STATUS=$(echo "$LOGS_RESPONSE" | tail -1 | cut -d: -f2)
LOGS_BODY=$(echo "$LOGS_RESPONSE" | head -n -1)

echo "📋 Status HTTP dos logs: $LOGS_HTTP_STATUS"

if [ "$LOGS_HTTP_STATUS" = "200" ]; then
    echo "✅ Logs recuperados com sucesso!"
    echo "📜 Resposta dos logs:"
    echo "$LOGS_BODY"
    
    # Extrair informações dos logs
    LOGS_COUNT=$(echo "$LOGS_BODY" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    logs_data = data.get('data', [])
    total = data.get('total', 0)
    print(f'Total: {total}, Logs na página: {len(logs_data)}')
    
    # Mostrar cada log
    for i, log in enumerate(logs_data):
        log_type = log.get('type', 'unknown')
        content = log.get('content', '')
        created_at = log.get('created_at', '')
        print(f'  Log {i+1}: [{log_type}] {content[:100]}...')
except Exception as e:
    print(f'Erro ao processar logs: {e}')
    pass
" 2>/dev/null)
    
    echo "📊 Detalhes dos logs:"
    echo "$LOGS_COUNT"
else
    echo "❌ Falha ao recuperar logs (status: $LOGS_HTTP_STATUS)"
    echo "   Resposta: $LOGS_BODY"
fi

# Cleanup se flag -d foi passada
if [ "$CLEANUP_FLAG" = true ]; then
    echo ""
    echo "🧹 Deletando código criado (flag -d ativa)..."
    
    DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/code/$CODE_ID" \
      --get \
      --data-urlencode "project_uuid=$PROJECT_UUID" \
      -w "\nHTTP_STATUS_CODE:%{http_code}")
    
    DELETE_HTTP_STATUS=$(echo "$DELETE_RESPONSE" | tail -1 | cut -d: -f2)
    DELETE_BODY=$(echo "$DELETE_RESPONSE" | head -n -1)
    
    echo "📋 Status HTTP da deleção: $DELETE_HTTP_STATUS"
    
    if [ "$DELETE_HTTP_STATUS" = "200" ] || [ "$DELETE_HTTP_STATUS" = "204" ]; then
        echo "✅ Código deletado com sucesso!"
    else
        echo "❌ Falha ao deletar código (status: $DELETE_HTTP_STATUS)"
        echo "   Resposta: $DELETE_BODY"
    fi
else
    echo ""
    echo "💡 Código criado mantido (ID: $CODE_ID)"
    echo "   Para deletar: $0 -d"
    echo "   Ou delete manualmente via API: DELETE $BASE_URL/code/$CODE_ID?project_uuid=$PROJECT_UUID"
fi

echo ""
echo "📊 Resumo do teste:"
echo "=================="
echo "   Aplicação respondendo: ✅"
echo "   Criação de código: $([ -n "$CODE_ID" ] && echo "✅" || echo "❌")"
echo "   Execução de código: $([ "$HTTP_STATUS" = "200" ] && echo "✅" || echo "❌")"
echo "   Listagem de logs: $([ "$LOGS_HTTP_STATUS" = "200" ] && echo "✅" || echo "❌")"
echo "   Limpeza de dados: $([ "$CLEANUP_FLAG" = true ] && echo "✅" || echo "➖")"

if [ "$HTTP_STATUS" != "200" ] || [ "$LOGS_HTTP_STATUS" != "200" ]; then
    echo ""
    echo "💡 Próximos passos para debug:"
    echo "   1. Verificar logs: docker logs codeactions-app-test"
    echo "   2. Debug completo: ./debug_application_crash.sh"
    echo "   3. Entrar no container: docker exec -it codeactions-app-test sh"
fi

echo ""
echo "🎯 Teste de logs concluído!"
