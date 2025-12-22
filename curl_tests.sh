#!/bin/bash

# ==============================================================================
# Testes de API - Weni Code Actions
# ==============================================================================

# Configurações
# HOST="https://code-actions.stg.cloud.weni.ai"
HOST="http://localhost:8050"
AUTH_TOKEN="Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJnTFNiNm93ZTF4WEN5QVN4c0NIeGxNdXNhcER0WXVxNTVOcTl4akc4X3FzIn0.eyJleHAiOjE3NjYxMzcwMTEsImlhdCI6MTc2NjA5MzgxMSwiYXV0aF90aW1lIjoxNzY2MDkzODA5LCJqdGkiOiJvZnJ0YWM6OTdlNWU3OGMtODJhOC02ZGM0LTNlNzItZWZjYjBkOTg3ZWI1IiwiaXNzIjoiaHR0cHM6Ly9hY2NvdW50cy53ZW5pLmFpL2F1dGgvcmVhbG1zL3dlbmktc3RhZ2luZyIsImF1ZCI6ImFjY291bnQiLCJzdWIiOiI5M2RkNjc0MC04MjczLTRkMjAtODNhZS05ZTU2NzE1ZmVmMWIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJ3ZW5pLXdlYmFwcCIsInNpZCI6ImM3NWEyYTU1LTU5MjItNGNhMi1hODRkLTVlMGE0ZTU4MmQxYSIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJvZmZsaW5lX2FjY2VzcyIsInVtYV9hdXRob3JpemF0aW9uIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwic2NvcGUiOiJlbWFpbCBwcm9maWxlIG9wZW5pZCBvZmZsaW5lX2FjY2VzcyIsImVtYWlsX3ZlcmlmaWVkIjpmYWxzZSwibmFtZSI6IlJhZmFlbCBTb2FyZXMiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJyYWZhZWwuc29hcmVzQHdlbmkuYWkiLCJnaXZlbl9uYW1lIjoiUmFmYWVsIiwiZmFtaWx5X25hbWUiOiJTb2FyZXMiLCJlbWFpbCI6InJhZmFlbC5zb2FyZXNAd2VuaS5haSJ9.fuPWdHB7bd-0XWFc3AZorFVX3_n18_p-CNLEUGahGYdocISlztxTmWWuznV74jjWiX_v43TapMJKUL9aDlmFWd3Vvkp-g_vExwWEhPhlLujdO5qT_OIj0n7QkPBz-FfIlussDr4RX-i0BVEX-JJW2grJ6OQfev7AYfLgOLgb1cRwsU85rP0wrZgiM2l1oDKWUGrIydtHEyTD08gD9zZJk4Md0dKTUAsipMwxOFFLGutGqbF-XzgOQYF7-2J37o4AoUuqoCQXoSJs-rbGFKFZ82ZA8GZUnWnqSGRMWWITmZ7FyPfwz6vkRf6BibdF07QZNJBIgm5-e-6HpLWcCC2j0w"
PROJECT_UUID="fa147fa6-5af0-4d99-9c00-043c89d97392"

echo "==== TESTANDO API CODE ACTIONS ===="
echo ""

# ==============================================================================
# 1. CRIAR CÓDIGO QUE CONTENHA LOGS
# ==============================================================================
echo "1. ✅ CRIANDO CÓDIGO COM LOGS..."

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

# Criar código Python com logs (tipo endpoint para compatibilidade com /action/endpoint/)
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

echo ""
echo "=============================================================================="
echo ""

# ==============================================================================
# 2. RODAR CÓDIGO
# ==============================================================================
echo "2. 🚀 EXECUTANDO CÓDIGO..."

if [ -z "$CODE_ID" ]; then
    echo "❌ CODE_ID não encontrado. Execute primeiro a criação do código."
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

echo ""
echo "=============================================================================="
echo ""

# wait 2 seconds
sleep 2
echo "2 segundos de espera..."


# ==============================================================================
# 3. CONSULTAR RUNS DO CÓDIGO EXECUTADO
# ==============================================================================
echo "3. 🔍 CONSULTANDO RUNS DO CÓDIGO..."

# Verificar se temos CODE_ID válido
if [ -z "$CODE_ID" ]; then
    echo "❌ CODE_ID não encontrado. Pulando consulta de runs."
else
    # Consultar por code_id (parâmetro obrigatório do endpoint)
    echo "📋 Listando todas as execuções por code_id: $CODE_ID"
    RUNS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
      -X GET "${HOST}/coderun?code_id=${CODE_ID}" \
      -H "Authorization: ${AUTH_TOKEN}")

HTTP_STATUS=$(echo "$RUNS_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
RUNS_BODY=$(echo "$RUNS_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Runs consultados com sucesso!"
    echo "Response: $RUNS_BODY"
    
    # Tentar extrair o primeiro RUN_ID da lista se não temos um
    if [ -z "$RUN_ID" ]; then
        RUN_ID=$(echo "$RUNS_BODY" | jq -r '.[0].id // empty')
        if [ ! -z "$RUN_ID" ] && [ "$RUN_ID" != "null" ]; then
            echo "🏃 Primeiro RUN_ID encontrado: $RUN_ID"
        else
            echo "ℹ️ Nenhuma execução encontrada na resposta"
        fi
    fi
else
    echo "❌ Erro ao consultar runs. HTTP Status: $HTTP_STATUS"
    echo "Response: $RUNS_BODY"
fi

# Se temos um RUN_ID específico, consultar detalhes
if [ ! -z "$RUN_ID" ]; then
    echo ""
    echo "🔍 Consultando detalhes do run específico..."
    RUN_DETAIL_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
      -X GET "${HOST}/coderun/${RUN_ID}" \
      -H "Authorization: ${AUTH_TOKEN}")
    
    HTTP_STATUS=$(echo "$RUN_DETAIL_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    RUN_DETAIL_BODY=$(echo "$RUN_DETAIL_RESPONSE" | sed '/HTTP_STATUS:/d')
    
    if [ "$HTTP_STATUS" = "200" ]; then
        echo "✅ Detalhes do run consultados com sucesso!"
        echo "Response: $RUN_DETAIL_BODY"
    else
        echo "❌ Erro ao consultar detalhes do run. HTTP Status: $HTTP_STATUS"
        echo "Response: $RUN_DETAIL_BODY"
    fi
fi
fi

echo ""
echo "=============================================================================="
echo ""

# ==============================================================================
# 4. CONSULTAR LOGS DO CÓDIGO EXECUTADO
# ==============================================================================
echo "4. 📜 CONSULTANDO LOGS DO CÓDIGO..."

# Verificar se temos CODE_ID válido
if [ -z "$CODE_ID" ]; then
    echo "❌ CODE_ID não encontrado. Pulando consulta de logs."
else
    # Consultar todos os logs por code_id
    echo "📋 Listando todos os logs por code_id: $CODE_ID"
    LOGS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
      -X GET "${HOST}/codelog?code_id=${CODE_ID}" \
      -H "Authorization: ${AUTH_TOKEN}")

HTTP_STATUS=$(echo "$LOGS_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
LOGS_BODY=$(echo "$LOGS_RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ Logs consultados com sucesso!"
    echo "Response: $LOGS_BODY"
    
    # Tentar extrair o primeiro LOG_ID se disponível
    LOG_ID=$(echo "$LOGS_BODY" | jq -r '.data[0].id // empty')
    if [ ! -z "$LOG_ID" ] && [ "$LOG_ID" != "null" ]; then
        echo "📜 Primeiro LOG_ID encontrado: $LOG_ID"
    else
        echo "ℹ️ Nenhum log encontrado - isso é normal se a execução ainda não gerou logs"
        
        # Verificar se há dados na resposta
        TOTAL_LOGS=$(echo "$LOGS_BODY" | jq -r '.total // 0')
        if [ "$TOTAL_LOGS" = "0" ]; then
            echo "💡 Total de logs: 0"
            echo "📝 NOTA: /action/endpoint/ executa código diretamente sem criar coderun/codelog"
            echo "    Para gerar logs consultáveis, use /run/ (requer correção do bug de argumentos)"
            echo "    Os logs do /action/endpoint/ são apenas no resultado da resposta direta"
        fi
    fi
else
    echo "❌ Erro ao consultar logs. HTTP Status: $HTTP_STATUS"
    echo "Response: $LOGS_BODY"
fi

# Se temos um LOG_ID específico, consultar detalhes
if [ ! -z "$LOG_ID" ]; then
    echo ""
    echo "🔍 Consultando detalhes do log específico..."
    LOG_DETAIL_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
      -X GET "${HOST}/codelog/${LOG_ID}" \
      -H "Authorization: ${AUTH_TOKEN}")
    
    HTTP_STATUS=$(echo "$LOG_DETAIL_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    LOG_DETAIL_BODY=$(echo "$LOG_DETAIL_RESPONSE" | sed '/HTTP_STATUS:/d')
    
    if [ "$HTTP_STATUS" = "200" ]; then
        echo "✅ Detalhes do log consultados com sucesso!"
        echo "Response: $LOG_DETAIL_BODY"
    else
        echo "❌ Erro ao consultar detalhes do log. HTTP Status: $HTTP_STATUS"
        echo "Response: $LOG_DETAIL_BODY"
    fi
fi
fi

echo ""
echo "=============================================================================="
echo ""

# ==============================================================================
# 5. DELETAR CÓDIGO E TODOS OS DADOS RELACIONADOS
# ==============================================================================
echo "5. 🗑️  DELETANDO CÓDIGO E DADOS RELACIONADOS..."

CLEANUP_FLAG=false
# verificar se a flag -d foi passada
if [ "$1" = "-d" ]; then
    CLEANUP_FLAG=true
fi

# Deletar apenas se a flag -d foi passada
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
echo "=============================================================================="
echo ""

# ==============================================================================
# COMANDOS EXTRAS PARA TESTE E DEBUG
# ==============================================================================
echo "🛠️ COMANDOS EXTRAS PARA DEBUG:"
echo ""

echo "# Verificar saúde da aplicação:"
echo "curl -X GET \"${HOST}/health\""
echo ""

echo "# Listar todos os códigos do projeto:"
echo "curl -X GET \"${HOST}/code?project_uuid=${PROJECT_UUID}\" \\"
echo "  -H \"Authorization: ${AUTH_TOKEN}\""
echo ""

echo "# Listar execuções de um código específico:"
echo "curl -X GET \"${HOST}/coderun?code_id=\${CODE_ID}\" \\"
echo "  -H \"Authorization: ${AUTH_TOKEN}\""
echo ""

echo "# Listar logs de um código específico:"
echo "curl -X GET \"${HOST}/codelog?code_id=\${CODE_ID}\" \\"
echo "  -H \"Authorization: ${AUTH_TOKEN}\""
echo ""

echo "# Consultar código específico:"
echo "curl -X GET \"${HOST}/code/\${CODE_ID}\" \\"
echo "  -H \"Authorization: ${AUTH_TOKEN}\""
echo ""

echo "# Criar código tipo ENDPOINT para testes de execução:"
echo "curl -X POST \"${HOST}/code?name=Endpoint%20Test&language=python&type=endpoint&project_uuid=${PROJECT_UUID}\" \\"
echo "  -H \"Authorization: ${AUTH_TOKEN}\" \\"
echo "  -H \"Content-Type: text/plain\" \\"
echo "  --data-raw 'def Run(engine): engine.result.set({\"message\": \"Hello from endpoint!\"}, content_type=\"json\")'"
echo ""

echo "# Executar via /run/ (apenas se corrigido o bug):"
echo "curl -X POST \"${HOST}/run/\${CODE_ID}\" \\"
echo "  -H \"Authorization: ${AUTH_TOKEN}\" \\"
echo "  -H \"Content-Type: application/json\""
echo ""

echo "# Executar endpoint (se for type=endpoint):"
echo "curl -X POST \"${HOST}/endpoint/\${CODE_ID}\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"test_data\": \"value\"}'"
echo ""

echo "# Executar action endpoint com rate limiting (usado no script):"
echo "curl -X POST \"${HOST}/action/endpoint/\${CODE_ID}?param=value\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -H \"X-Custom-Header: value\" \\"
echo "  -d '{\"action_data\": \"value\"}'"
echo ""

echo "# MÉTODO USADO NO SCRIPT (recomendado):"
echo "# - /action/endpoint/ funciona com todos os tipos de código"
echo "# - Não tem bug de argumentos vazios"
echo "# - Inclui rate limiting automático"
echo "# - Suporte completo a headers, params e body"
echo ""


echo "==== TESTES CONCLUÍDOS ====" 
echo ""
echo "📋 RESUMO DOS MÉTODOS DE EXECUÇÃO:"
echo "✅ /action/endpoint/ - USADO NO SCRIPT (recomendado)"
echo "   - Funciona com todos os tipos de código"
echo "   - Sem bug de argumentos vazios"
echo "   - Rate limiting integrado"
echo "   - Suporte completo a headers/params/body"
echo ""
echo "⚠️  /run/ - TEM BUG com argumentos vazios"
echo "   - Falha quando headers estão vazios"
echo "   - Só funciona perfeitamente com headers/params válidos"
echo ""
echo "✅ /endpoint/ - Funciona apenas com type=endpoint"
echo "   - Sem bug de argumentos"
echo "   - Limitado a códigos tipo endpoint"
echo ""
echo "📋 PARÂMETROS DOS ENDPOINTS DE CONSULTA:"
echo "• /code - usa project_uuid (listar códigos do projeto)"
echo "• /coderun - usa code_id (listar execuções de um código)"
echo "• /codelog - usa code_id OU run_id (listar logs)"
echo ""
echo "📝 SOBRE LOGS E EXECUÇÕES:"
echo "• /action/endpoint/ - Execução direta, logs só na resposta HTTP"
echo "• /run/ - Cria coderun + codelog consultáveis (mas tem bug de args vazios)"  
echo "• /endpoint/ - Similar ao /run/ mas só para type=endpoint"
echo ""
echo "💡 Para logs persistidos e consultáveis:"
echo "   1. Use type=endpoint + /endpoint/ (sem bugs)"
echo "   2. Ou corrija o bug do /run/ no código Go (service.go:155)"
echo ""
echo "🧪 TESTES DE INTEGRAÇÃO S3:"
echo "   Para testar a listagem de logs no S3/LocalStack:"
echo "   ./start_localstack_for_tests.sh  # Iniciar LocalStack"
echo "   ./test_s3_integration.sh         # Executar testes"
echo "   Veja: S3_INTEGRATION_TESTS.md"
echo ""
echo "🎯 BUG S3 TIMEZONE IDENTIFICADO E CORRIGIDO:"
echo "   PROBLEMA: Logs salvos em UTC (Python) mas buscados em timezone local (Go)"
echo "   SINTOMA:  API retorna {\"data\":[]} mesmo com logs existindo no S3"
echo "   CAUSA:    time.Now() vs time.Now().UTC() -> diferença de 1 dia"
echo "   SOLUÇÃO:  Forçar UTC na busca (internal/codelog/s3/codelog.go)"
echo ""
echo "   📂 Exemplo:"
echo "      Logs salvos em: /logs/2025/12/18/{run_id}/"
echo "      Go buscava em:  /logs/2025/12/17/{run_id}/ ❌"
echo "      Go busca em:    /logs/2025/12/18/{run_id}/ ✅"
echo ""
