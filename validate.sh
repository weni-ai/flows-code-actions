HOST="https://code-actions.stg.cloud.weni.ai"
AUTH_TOKEN="Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJnTFNiNm93ZTF4WEN5QVN4c0NIeGxNdXNhcER0WXVxNTVOcTl4akc4X3FzIn0.eyJleHAiOjE3Njg1MjQzNDUsImlhdCI6MTc2ODQ4MTE0NSwiYXV0aF90aW1lIjoxNzY4NDgxMTQzLCJqdGkiOiJvZnJ0YWM6ZmQ3MWY2MDQtN2NlYi04ZDgxLWExNmEtMDY2NjQ4ODVmMGY3IiwiaXNzIjoiaHR0cHM6Ly9hY2NvdW50cy53ZW5pLmFpL2F1dGgvcmVhbG1zL3dlbmktc3RhZ2luZyIsImF1ZCI6ImFjY291bnQiLCJzdWIiOiI5M2RkNjc0MC04MjczLTRkMjAtODNhZS05ZTU2NzE1ZmVmMWIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJ3ZW5pLXdlYmFwcCIsInNpZCI6IjNmYzA2MTUwLTRiZjktNGQyYi1hOWZiLWQyMzdmMjA0YzNkMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJvZmZsaW5lX2FjY2VzcyIsInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwic2NvcGUiOiJlbWFpbCBwcm9maWxlIG9wZW5pZCBvZmZsaW5lX2FjY2VzcyIsImVtYWlsX3ZlcmlmaWVkIjpmYWxzZSwibmFtZSI6IlJhZmFlbCBTb2FyZXMiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJyYWZhZWwuc29hcmVzQHdlbmkuYWkiLCJnaXZlbl9uYW1lIjoiUmFmYWVsIiwibG9jYWxlIjoiZW4iLCJmYW1pbHlfbmFtZSI6IlNvYXJlcyIsImVtYWlsIjoicmFmYWVsLnNvYXJlc0B3ZW5pLmFpIn0.DiiumVecDjQ82PXcXceW-mNZFdK1wBSE4GkBHP-xcY4_x1VtunebalOVhjzWXvzXmgUtS42erBayy6vfcPKl6SLAqlWNdj7GGPNIV4S2QdALmolPgSXwRQt0wka-0LV2IRuqIzeXV1hZwOAOG-YLA0XSOiucfr_amBD9_0PDbrtN8o3wEbWWhExf3WbpFKznbx1mqaVoSTdGxPBJnHE1lcUDfR4fbxZsDE9UScTLpYWikVaqip35pWw9v5EXgCo0kkeEX4aPJMPekCyLlPgtYmV-al-HJ1DZ1raUwdfR8N8KzaZykhLFID_msBF5FhcJvuRyNv1fMV-FPT-rl9G9_A"
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
  -X POST "${HOST}/code?name=Código%20com%20Logs&language=python&type=endpoint&project_uuid=${PROJECT_UUID}" \
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
    echo "✅ Código executado com sucesso via /action/endpoint/!"
    echo "Response: $EXECUTION_BODY"
    echo "📊 Resultado da execução detalhado abaixo:"
    echo "$EXECUTION_BODY" | jq . 2>/dev/null || echo "$EXECUTION_BODY"
    
    # Para action/endpoint, o resultado é direto (não retorna RUN_ID explícito)
    echo "ℹ️ Execução via /action/endpoint/ concluída - logs foram gerados automaticamente"
else
    echo "❌ Erro ao executar código via /action/endpoint/. HTTP Status: $HTTP_STATUS"
    echo "Response: $EXECUTION_BODY"
    
    if [ "$HTTP_STATUS" = "404" ]; then
        echo "💡 Dica: Códigos type=flow podem não funcionar com /action/endpoint/"
        echo "        Teste criar um código type=endpoint para este endpoint"
    fi
fi
