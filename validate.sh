# HOST="https://code-actions.stg.cloud.weni.ai"
# AUTH_TOKEN="PLACE YOUR TOKEN HERE"
# PROJECT_UUID="fa147fa6-5af0-4d99-9c00-043c89d97392"

HOST="http://localhost:8050"
AUTH_TOKEN="PLACE YOUR TOKEN HERE"

// generate a uuid
PROJECT_UUID=$(uuidgen)

SECRET_NAME="MY_API_KEY"
SECRET_VALUE="super-secret-value-123"

# ---------------------------------------------------------------------------
# 1. Criar secret
# ---------------------------------------------------------------------------
echo "🔑 Criando secret '${SECRET_NAME}'..."
SECRET_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X POST "${HOST}/secret" \
  -H "Authorization: ${AUTH_TOKEN}" \
  -H "Content-Type: application/json" \
  --data-raw "{\"name\": \"${SECRET_NAME}\", \"value\": \"${SECRET_VALUE}\", \"project_uuid\": \"${PROJECT_UUID}\"}")

HTTP_STATUS=$(echo "$SECRET_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
SECRET_BODY=$(echo "$SECRET_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "201" ]; then
    echo "✅ Secret criado com sucesso!"
    SECRET_ID=$(echo "$SECRET_BODY" | jq -r '.id')
    echo "💾 SECRET_ID: $SECRET_ID"
else
    echo "❌ Erro ao criar secret. HTTP Status: $HTTP_STATUS"
    echo "Response: $SECRET_BODY"
    exit 1
fi

# ---------------------------------------------------------------------------
# 2. Criar código Python que usa o secret via engine.secrets
# ---------------------------------------------------------------------------
PYTHON_CODE_WITH_LOGS='import json
from datetime import datetime

def Run(engine):
    """
    Demonstra o uso de secrets carregados automaticamente pelo engine.
    """

    engine.log.info("🚀 Iniciando execução do código")

    try:
        engine.log.info("📊 Processando dados de entrada...")

        data = {
            "timestamp": datetime.now().isoformat(),
            "status": "processing",
            "user_id": 12345,
            "operation": "data_analysis",
            "request_params": dict(engine.params.items()) if engine.params else {},
            "request_body": engine.body if engine.body else "No body"
        }

        engine.log.info(f"📋 Dados processados: {json.dumps(data, indent=2)}")

        # --- Uso de secrets ---
        available_secrets = engine.secrets.keys()
        engine.log.info(f"🔐 Secrets disponíveis: {list(available_secrets)}")
        api_key = engine.secrets.get("MY_API_KEY", "chave-nao-configurada")
        engine.log.info(f"🔑 MY_API_KEY (preview): {api_key[:8]}...")
        # Simular alguma lógica de negócio
        processed_items = 0
        for i in range(5):
            processed_items += 30
            engine.log.debug(f"⚙️  Processando lote {i+1}: {processed_items} itens processados")

        result_data = {
            "status": "success",
            "message": "Código executado com sucesso e logs gerados!",
            "processed_items": processed_items,
            "execution_time": datetime.now().isoformat(),
            "secrets_loaded": len(list(available_secrets)),
            "api_key_preview": api_key[:8] + "..." if len(api_key) > 8 else api_key,
            "engine_info": {
                "has_params": bool(engine.params),
                "has_body": bool(engine.body),
                "has_headers": bool(engine.header)
            }
        }

        engine.log.info("✅ Processamento concluído com sucesso")
        engine.result.set(result_data, status_code=200, content_type="json")
        engine.log.info("💾 Resultado salvo com sucesso")

    except Exception as e:
        error_msg = f"❌ Erro durante execução: {str(e)}"
        engine.log.error(error_msg)
        engine.result.set(
            {"status": "error", "message": str(e), "timestamp": datetime.now().isoformat()},
            status_code=500,
            content_type="json"
        )

    finally:
        engine.log.info("🏁 Finalizando execução do código")'

# ---------------------------------------------------------------------------
# 3. Criar o código no servidor
# ---------------------------------------------------------------------------
RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X POST "${HOST}/code?name=codetest&language=python&type=endpoint&project_uuid=${PROJECT_UUID}" \
  -H "Authorization: ${AUTH_TOKEN}" \
  -H "Content-Type: text/plain" \
  --data-raw "${PYTHON_CODE_WITH_LOGS}")

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
RESPONSE_BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "201" ]; then
    echo "✅ Código criado com sucesso!"
    echo "Response: $RESPONSE_BODY"
    CODE_ID=$(echo "$RESPONSE_BODY" | jq -r '.id')
    echo "💾 CODE_ID salvo: $CODE_ID"
else
    echo "❌ Erro ao criar código. HTTP Status: $HTTP_STATUS"
    echo "Response: $RESPONSE_BODY"
    exit 1
fi

# ---------------------------------------------------------------------------
# 4. Executar via action/endpoint
# ---------------------------------------------------------------------------
echo "🚀 Executando via /action/endpoint/ (bypassa bugs do coderunner)..."
EXECUTION_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X POST "${HOST}/action/endpoint/${CODE_ID}?test_param=action_value&execution_mode=manual" \
  -H "Content-Type: application/json" \
  -H "X-Test-Header: action-execution" \
  -H "X-Execution-Source: curl-test" \
  --data-raw '{"test_data": "action_endpoint_test", "timestamp": "'$(date -Iseconds)'", "source": "curl_script"}')

HTTP_STATUS=$(echo "$EXECUTION_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
EXECUTION_BODY=$(echo "$EXECUTION_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "📊 Resultado da execução:"
    echo "$EXECUTION_BODY" | jq . 2>/dev/null || echo "$EXECUTION_BODY"
else
    echo "❌ Erro ao executar código via /action/endpoint/. HTTP Status: $HTTP_STATUS"
    echo "Response: $EXECUTION_BODY"
    exit 1
fi

# ---------------------------------------------------------------------------
# 5. Buscar a execução mais recente (coderun) para obter o run_id
# ---------------------------------------------------------------------------
echo ""
echo "🔍 Buscando execução mais recente para o código..."

# Aguarda um momento para garantir que os logs foram persistidos
sleep 2

CODERUN_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  "${HOST}/coderun?code_id=${CODE_ID}" \
  -H "Authorization: ${AUTH_TOKEN}")

HTTP_STATUS=$(echo "$CODERUN_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
CODERUN_BODY=$(echo "$CODERUN_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" != "200" ]; then
    echo "❌ Erro ao buscar coderuns. HTTP Status: $HTTP_STATUS"
    echo "Response: $CODERUN_BODY"
    exit 1
fi

# Pega o run_id da execução mais recente (último elemento da lista)
RUN_ID=$(echo "$CODERUN_BODY" | jq -r '.[-1].id // empty')

if [ -z "$RUN_ID" ]; then
    echo "⚠️  Nenhuma execução encontrada para code_id=${CODE_ID}"
    exit 1
fi

echo "✅ RUN_ID encontrado: $RUN_ID"

# ---------------------------------------------------------------------------
# 6. Consultar os logs da execução
# ---------------------------------------------------------------------------
echo ""
echo "📋 Consultando logs da execução (run_id=${RUN_ID})..."

LOGS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  "${HOST}/codelog?run_id=${RUN_ID}&code_id=${CODE_ID}&page=1" \
  -H "Authorization: ${AUTH_TOKEN}")

HTTP_STATUS=$(echo "$LOGS_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
LOGS_BODY=$(echo "$LOGS_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" != "200" ]; then
    echo "❌ Erro ao buscar logs. HTTP Status: $HTTP_STATUS"
    echo "Response: $LOGS_BODY"
    exit 1
fi

TOTAL_LOGS=$(echo "$LOGS_BODY" | jq '.total')
echo "📊 Total de logs: $TOTAL_LOGS"
echo ""
echo "📝 Logs da execução:"
echo "$LOGS_BODY" | jq -r '.data[] | "[\(.type | ascii_upcase)] \(.content)"' 2>/dev/null \
    || echo "$LOGS_BODY" | jq .

