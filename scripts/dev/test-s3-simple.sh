#!/bin/bash

# Script simplificado para testar S3 com LocalStack

set -e

ENDPOINT="http://localhost:4566"
BUCKET="codeactions-dev"

echo "🧪 Testando operações S3 no LocalStack (modo simplificado)..."

# Verifica se LocalStack está rodando
if ! curl -s $ENDPOINT/_localstack/health > /dev/null; then
    echo "❌ LocalStack não está rodando. Execute primeiro: ./scripts/start-localstack.sh"
    exit 1
fi

echo "✅ LocalStack está rodando"

# Testa listar buckets
echo "📋 Listando buckets..."
aws --endpoint-url=$ENDPOINT s3 ls

# Cria um arquivo simples para teste
echo "📄 Criando arquivo de teste..."
echo -n '{"test": "hello from localstack"}' > /tmp/simple-test.json

# Testa upload com curl (mais compatível)
echo "⬆️  Fazendo upload com curl..."
curl -X PUT \
    -H "Content-Type: application/json" \
    -d @/tmp/simple-test.json \
    "$ENDPOINT/$BUCKET/codeactions/logs/test-simple.json"

echo ""
echo "📂 Listando objetos no bucket..."
aws --endpoint-url=$ENDPOINT s3 ls s3://$BUCKET/ --recursive

# Testa download com curl  
echo "⬇️  Fazendo download com curl..."
curl -s "$ENDPOINT/$BUCKET/codeactions/logs/test-simple.json" > /tmp/downloaded-simple.json

# Compara arquivos
echo "🔍 Comparando arquivos..."
if diff /tmp/simple-test.json /tmp/downloaded-simple.json > /dev/null; then
    echo "✅ Teste passou! Upload/Download funcionando corretamente."
else
    echo "❌ Teste falhou! Arquivos são diferentes."
    echo "Original:"
    cat /tmp/simple-test.json
    echo "Downloaded:"
    cat /tmp/downloaded-simple.json
    exit 1
fi

# Limpeza
rm -f /tmp/simple-test.json /tmp/downloaded-simple.json

echo ""
echo "🎉 LocalStack S3 está funcionando corretamente!"
echo ""
echo "🚀 Configure as variáveis e execute sua aplicação:"
echo "   source config.localstack.example"
echo "   air -d"
