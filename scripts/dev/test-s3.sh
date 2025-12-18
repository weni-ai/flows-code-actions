#!/bin/bash

# Script para testar operações S3 com LocalStack

set -e

ENDPOINT="http://localhost:4566"
BUCKET="codeactions-dev"

echo "🧪 Testando operações S3 no LocalStack..."

# Verifica se LocalStack está rodando
if ! curl -s $ENDPOINT/health > /dev/null; then
    echo "❌ LocalStack não está rodando. Execute primeiro: ./scripts/start-localstack.sh"
    exit 1
fi

echo "✅ LocalStack está rodando"

# Testa listar buckets
echo "📋 Listando buckets..."
aws --endpoint-url=$ENDPOINT s3 ls

# Testa criar um arquivo de teste
echo "📄 Criando arquivo de teste..."
TEST_FILE="/tmp/test-codelog.json"
cat > $TEST_FILE << EOF
{
    "id": "test-log-id",
    "runId": "test-run-id", 
    "codeId": "test-code-id",
    "type": "info",
    "content": "Teste de log para LocalStack",
    "createdAt": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)",
    "updatedAt": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)"
}
EOF

# Upload do arquivo usando s3api put-object com configuração para evitar problemas de trailer
echo "⬆️  Fazendo upload do arquivo..."
AWS_CLI_FILE_ENCODING=UTF-8 \
aws --endpoint-url=$ENDPOINT \
    --cli-connect-timeout 60 \
    --cli-read-timeout 60 \
    s3api put-object \
    --bucket $BUCKET \
    --key codeactions/logs/2024/12/10/test-run-id/test-code-id/test-log-id.json \
    --body $TEST_FILE \
    --content-type application/json \
    --metadata test=true

# Lista arquivos no bucket
echo "📂 Listando arquivos no bucket..."
aws --endpoint-url=$ENDPOINT s3 ls s3://$BUCKET/codeactions/logs/ --recursive

# Download do arquivo usando s3api get-object
echo "⬇️  Fazendo download do arquivo..."
aws --endpoint-url=$ENDPOINT s3api get-object \
    --bucket $BUCKET \
    --key codeactions/logs/2024/12/10/test-run-id/test-code-id/test-log-id.json \
    /tmp/downloaded-test.json

# Compara arquivos
echo "🔍 Comparando arquivos..."
if diff $TEST_FILE /tmp/downloaded-test.json > /dev/null; then
    echo "✅ Teste passou! Upload/Download funcionando corretamente."
else
    echo "❌ Teste falhou! Arquivos são diferentes."
    exit 1
fi

# Limpeza
rm -f $TEST_FILE /tmp/downloaded-test.json

echo ""
echo "🎉 Todos os testes passaram! LocalStack S3 está funcionando corretamente."
echo ""
echo "🏃 Para usar com sua aplicação, configure as variáveis:"
echo "   source config.localstack.example"

