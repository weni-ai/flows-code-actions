#!/bin/bash

# Teste mínimo para verificar execução de código Python
echo "🧪 Teste Mínimo - Execução de Código"
echo "===================================="

BASE_URL="http://localhost:8050"
PROJECT_UUID="test-debug-$(date +%s)"

echo "🔍 Verificando se aplicação está respondendo..."
if ! curl -s -f "$BASE_URL/health" &>/dev/null; then
    echo "❌ Aplicação não está respondendo em $BASE_URL"
    echo "   Execute: ./run_image_tests.sh --debug-crash"
    exit 1
fi

echo "✅ Aplicação está respondendo"

echo ""
echo "🐍 Testando código Python extremamente simples..."

# Código Python mais simples possível
SIMPLE_CODE='def Run(engine):
    print("Hello from Python")
    engine.result.set({"status": "ok", "message": "success"}, content_type="json")'

echo "📝 Código de teste:"
echo "$SIMPLE_CODE"

echo ""
echo "📤 Criando código via API..."

# Criar código
RESPONSE=$(curl -s -X POST "$BASE_URL/code" \
  -H "Content-Type: text/plain" \
  -d "$SIMPLE_CODE" \
  --get \
  --data-urlencode "project_uuid=$PROJECT_UUID" \
  --data-urlencode "name=Debug Test" \
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
echo "🚀 Executando código..."

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
    echo ""
    echo "🔍 Para investigar:"
    echo "   docker logs codeactions-app-test --tail 50"
    echo "   ./debug_application_crash.sh"
fi

echo ""
echo "📊 Resumo do teste:"
echo "=================="
echo "   Aplicação respondendo: ✅"
echo "   Criação de código: $([ -n "$CODE_ID" ] && echo "✅" || echo "❌")"
echo "   Execução de código: $([ "$HTTP_STATUS" = "200" ] && echo "✅" || echo "❌")"

if [ "$HTTP_STATUS" != "200" ]; then
    echo ""
    echo "💡 Próximos passos:"
    echo "   1. Verificar logs: docker logs codeactions-app-test"
    echo "   2. Debug completo: ./debug_application_crash.sh"
    echo "   3. Entrar no container: docker exec -it codeactions-app-test sh"
fi 