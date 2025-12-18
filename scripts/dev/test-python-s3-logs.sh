#!/bin/bash

# Script para testar logs Python com S3

set -e

cd /home/rafaelsoares/weni/weni-ai/codeactions

echo "🧪 Testando logs Python com S3..."

# Verifica se LocalStack está rodando
if ! curl -s http://localhost:4566/_localstack/health > /dev/null; then
    echo "❌ LocalStack não está rodando. Execute:"
    echo "   ./scripts/dev/start-localstack.sh"
    exit 1
fi

echo "✅ LocalStack está rodando"

# Configurar variáveis de ambiente para S3
export FLOWS_CODE_ACTIONS_S3_ENABLED=true
export FLOWS_CODE_ACTIONS_S3_ENDPOINT=http://localhost:4566
export FLOWS_CODE_ACTIONS_S3_REGION=us-east-1
export FLOWS_CODE_ACTIONS_S3_BUCKET_NAME=codeactions-dev
export FLOWS_CODE_ACTIONS_S3_PREFIX=codeactions
export FLOWS_CODE_ACTIONS_S3_ACCESS_KEY_ID=test
export FLOWS_CODE_ACTIONS_S3_SECRET_ACCESS_KEY=test

# MongoDB configs (para Result, não afeta logs)
export FLOWS_CODE_ACTIONS_MONGO_DB_URI=mongodb://localhost:27017
export FLOWS_CODE_ACTIONS_MONGO_DB_NAME=code-actions-test

echo "📋 Configurações:"
echo "   S3 Enabled: $FLOWS_CODE_ACTIONS_S3_ENABLED"
echo "   S3 Endpoint: $FLOWS_CODE_ACTIONS_S3_ENDPOINT"
echo "   S3 Bucket: $FLOWS_CODE_ACTIONS_S3_BUCKET_NAME"

# Criar código Python de teste que gera logs
echo "📝 Criando código de teste..."
cat > /tmp/test_action.py << 'EOF'
def Run(engine):
    # Teste logs em S3
    engine.log.info("Teste de log INFO no S3")
    engine.log.debug("Teste de log DEBUG no S3") 
    engine.log.error("Teste de log ERROR no S3")
    
    # Log com conteúdo longo
    long_content = "Este é um teste de log longo. " * 100
    engine.log.info(f"Log longo: {long_content}")
    
    # Resultado
    engine.result.set({
        "message": "Logs enviados para S3 com sucesso!",
        "timestamp": "2024-12-15T22:00:00Z"
    }, content_type="json")
EOF

# Copiar arquivo de teste para engines/py/
cp /tmp/test_action.py engines/py/action.py

echo "🐍 Executando engine Python..."

# IDs de teste  
TEST_RUN_ID="67b5551d92d1ff6471e94991"
TEST_CODE_ID="67b5551d92d1ff6471e94992"

# Executar main.py com logs habilitados para S3
cd engines/py
python main.py -r $TEST_RUN_ID -c $TEST_CODE_ID -a '{}' -b '' -H '{}'

echo ""
echo "📂 Verificando se logs foram salvos no S3..."

# Listar objetos no bucket para verificar se logs foram criados
aws --endpoint-url=http://localhost:4566 s3 ls s3://codeactions-dev/codeactions/logs/ --recursive | tail -10

echo ""
echo "🔍 Verificando estrutura dos logs no S3..."

# Pegar o log mais recente e mostrar conteúdo
LATEST_LOG=$(aws --endpoint-url=http://localhost:4566 s3 ls s3://codeactions-dev/codeactions/logs/ --recursive | tail -1 | awk '{print $4}')

if [ -n "$LATEST_LOG" ]; then
    echo "📄 Conteúdo do log mais recente:"
    aws --endpoint-url=http://localhost:4566 s3 cp s3://codeactions-dev/$LATEST_LOG /tmp/latest-log.json
    cat /tmp/latest-log.json | jq '.'
else
    echo "❌ Nenhum log encontrado no S3"
    exit 1
fi

echo ""
echo "✅ Teste concluído!"
echo "🎉 Python engine agora salva logs no S3 seguindo o mesmo padrão do Go!"

# Limpeza
rm -f /tmp/test_action.py /tmp/latest-log.json
cd ../..
