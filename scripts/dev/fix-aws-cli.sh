#!/bin/bash

# Script para configurar AWS CLI para funcionar corretamente com LocalStack
# Resolve o problema do "x-amz-trailer header is not supported"

set -e

echo "🔧 Configurando AWS CLI para LocalStack..."

# Criar diretório de configuração do AWS CLI se não existir
mkdir -p ~/.aws

# Configuração do AWS CLI para LocalStack
cat > ~/.aws/config << 'EOF'
[default]
region = us-east-1
output = json
# Configurações específicas para LocalStack
s3 =
    max_concurrent_requests = 1
    max_bandwidth = 50MB/s
    multipart_threshold = 64MB
    multipart_chunksize = 16MB
    max_queue_size = 1000
    use_accelerate_endpoint = false
    addressing_style = path
EOF

# Configuração de credenciais para LocalStack
cat > ~/.aws/credentials << 'EOF'
[default]
aws_access_key_id = test
aws_secret_access_key = test
EOF

echo "✅ Configuração AWS CLI criada!"
echo ""
echo "📋 Configurações aplicadas:"
echo "   - Região: us-east-1"
echo "   - Threshold multipart: 64MB (evita x-amz-trailer para arquivos pequenos)"
echo "   - Path addressing style (compatível com LocalStack)"
echo "   - Credenciais de teste configuradas"
echo ""
echo "🧪 Testando configuração..."

# Teste básico
if aws --endpoint-url=http://localhost:4566 s3 ls > /dev/null 2>&1; then
    echo "✅ AWS CLI configurado corretamente para LocalStack!"
else
    echo "❌ Ainda há problemas na configuração"
fi

echo ""
echo "💡 Para usar:"
echo "   aws --endpoint-url=http://localhost:4566 s3api put-object ..."
echo "   Ou execute: ./scripts/dev/test-s3-final.sh"
