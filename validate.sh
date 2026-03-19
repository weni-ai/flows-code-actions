HOST="https://code-actions.stg.cloud.weni.ai"
AUTH_TOKEN="PLACE YOUR TOKEN HERE"
PROJECT_UUID="fa147fa6-5af0-4d99-9c00-043c89d97392"

PYTHON_CODE_WITH_LOGS='import json
from datetime import datetime

def Run(engine):
    """
    Função principal que será executada pelo engine
    O engine fornece objetos para log e resultado
    """
    
    # Usar o sistema de logs integrado do engine
    engine.log.info("🚀 Iniciando execução do código")
    
    try:
        # Simular processamento com logs detalhados
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
        
        # Simular processamento com diferentes tipos de log
        engine.log.debug("🔍 Executando validação de dados...")
        
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
            "logs_generated": True,
            "engine_info": {
                "has_params": bool(engine.params),
                "has_body": bool(engine.body),
                "has_headers": bool(engine.header)
            }
        }
        
        engine.log.info("✅ Processamento concluído com sucesso")
        engine.log.debug(f"📤 Resultado detalhado: {json.dumps(result_data, indent=2)}")
        
        # Definir resultado usando o engine
        engine.result.set(result_data, status_code=200, content_type="json")
        
        engine.log.info("💾 Resultado salvo com sucesso")
        
    except Exception as e:
        error_msg = f"❌ Erro durante execução: {str(e)}"
        engine.log.error(error_msg)
        
        error_result = {
            "status": "error", 
            "message": str(e),
            "timestamp": datetime.now().isoformat()
        }
        
        engine.result.set(error_result, status_code=500, content_type="json")
        
    finally:
        engine.log.info("🏁 Finalizando execução do código")'


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


# Executar via action/endpoint (funciona com qualquer tipo de código)
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
    echo "Response: $EXECUTION_BODY"
    echo "📊 Resultado da execução detalhado abaixo:"
    echo "$EXECUTION_BODY" | jq . 2>/dev/null || echo "$EXECUTION_BODY"
    
    # Para action/endpoint, o resultado é direto (não retorna RUN_ID explícito)
    echo "ℹ️ Execução via /action/endpoint/ concluída - logs foram gerados automaticamente"
else
    echo "❌ Erro ao executar código via /action/endpoint/. HTTP Status: $HTTP_STATUS"
    echo "Response: $EXECUTION_BODY"
fi
