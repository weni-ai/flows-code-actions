# Troubleshooting - Testes de Imagem Docker

Este documento contém soluções para problemas comuns encontrados ao executar os testes de integração.

## 🔴 **Erro: MongoDB Connection Refused**

### **Sintoma:**
```
time="2025-06-06T19:49:00Z" level=fatal msg="mongodb fail to ping: server selection error: context deadline exceeded, current topology: { Type: Unknown, Servers: [{ Addr: localhost:27017, Type: Unknown, Last error: dial tcp [::1]:27017: connect: connection refused }, ] }"
```

### **Causa:**
A aplicação está tentando conectar no MongoDB via `localhost:27017`, mas dentro do container, `localhost` refere-se ao próprio container da aplicação, não ao container do MongoDB.

### **Soluções:**

#### **1. Verificar Configuração da Rede Docker**
```bash
# Verificar se containers estão na mesma rede
./debug_containers.sh

# Ou manualmente:
docker network inspect codeactions-test-network
```

#### **2. Verificar Variáveis de Ambiente**
A aplicação deve usar os nomes dos containers como hostnames:
```bash
# Correto (usado no script):
FLOWS_CODE_ACTIONS_MONGO_DB_URI="mongodb://mongo-test:27017"
FLOWS_CODE_ACTIONS_MONGO_DB_NAME="codeactions_test"
FLOWS_CODE_ACTIONS_REDIS="redis://redis-test:6379/1"

# Incorreto (não funciona em containers):
FLOWS_CODE_ACTIONS_MONGO_DB_URI="mongodb://localhost:27017"
```

#### **3. Executar com Debug Completo**
```bash
# Executar com logs detalhados
./run_image_tests.sh --verbose --no-cleanup --logs

# Em caso de falha, verificar conectividade
./debug_containers.sh
```

#### **4. Limpeza e Restart Completo**
```bash
# Limpar tudo e começar do zero
docker stop $(docker ps -q) 2>/dev/null || true
docker system prune -f
./run_image_tests.sh --build
```

## 🔴 **Erro: Container da Aplicação Para**

### **Sintoma:**
```
❌ Container da aplicação parou de funcionar!
```

### **Diagnóstico:**
```bash
# Ver logs da aplicação
docker logs codeactions-app-test --tail 50

# Ver status do container
docker ps -a --filter name=codeactions-app-test
```

### **Soluções:**

#### **1. Problemas de Configuração**
Verificar se as variáveis de ambiente estão corretas:
```bash
docker exec codeactions-app-test env | grep FLOWS_CODE_ACTIONS
```

#### **2. Problemas de Imagem**
Rebuildar a imagem:
```bash
./run_image_tests.sh --build
```

#### **3. Problemas de Porta**
Verificar se a porta não está em uso:
```bash
netstat -tuln | grep 8050
# ou
lsof -i :8050
```

## 🔴 **Erro: Health Check Falha**

### **Sintoma:**
```
❌ Aplicação não ficou disponível em http://localhost:8050 após X tentativas
```

### **Diagnóstico:**
```bash
# Testar manualmente
curl -v http://localhost:8050/health

# Verificar se aplicação está escutando na porta correta
docker exec codeactions-app-test netstat -tuln
```

### **Soluções:**

#### **1. Aguardar Mais Tempo**
A aplicação pode demorar para inicializar:
```bash
# O script já aguarda 60 tentativas (2 minutos)
# Se necessário, aumentar timeout no script
```

#### **2. Verificar Logs de Inicialização**
```bash
docker logs codeactions-app-test -f
```

#### **3. Testar Conectividade Interna**
```bash
# Entrar no container e testar localmente
docker exec -it codeactions-app-test sh
wget -O- http://localhost:8050/health
```

## 🔴 **Erro: Dependências Não Ficam Prontas**

### **Sintoma:**
```
❌ MongoDB não ficou pronto
❌ Redis não ficou pronto
```

### **Soluções:**

#### **1. Verificar Recursos do Sistema**
```bash
# Verificar uso de CPU/memória
docker stats

# Verificar espaço em disco
df -h
```

#### **2. Aguardar Mais Tempo**
```bash
# Containers podem demorar em sistemas lentos
# O script já aguarda 30 tentativas por serviço
```

#### **3. Verificar Imagens**
```bash
# Atualizar imagens base
docker pull mongo:7
docker pull redis:7-alpine
```

## 🛠️ **Scripts de Debug**

### **Script Principal com Debug:**
```bash
./run_image_tests.sh --verbose --no-cleanup --logs
```

### **Script de Diagnóstico:**
```bash
./debug_containers.sh
```

### **Limpeza Completa:**
```bash
# Parar todos os containers
docker stop $(docker ps -q) 2>/dev/null || true

# Remover containers de teste
docker rm codeactions-app-test mongo-test redis-test 2>/dev/null || true

# Remover rede
docker network rm codeactions-test-network 2>/dev/null || true

# Limpeza geral
docker system prune -f
```

## 🔍 **Verificações Manuais**

### **1. Testar Conectividade Entre Containers:**
```bash
# Ping entre containers
docker exec codeactions-app-test ping -c 2 mongo-test
docker exec codeactions-app-test ping -c 2 redis-test

# Testar portas específicas
docker exec codeactions-app-test telnet mongo-test 27017
docker exec codeactions-app-test telnet redis-test 6379
```

### **2. Verificar DNS Resolution:**
```bash
docker exec codeactions-app-test nslookup mongo-test
docker exec codeactions-app-test nslookup redis-test
```

### **3. Testar Aplicação Diretamente:**
```bash
# Health check
curl http://localhost:8050/health

# Criar código via API
curl -X POST "http://localhost:8050/code?project_uuid=test&name=test&type=endpoint&language=python" \
  -d 'def Run(engine): engine.result.set({"test": "ok"}, content_type="json")'
```

## 📞 **Quando Pedir Ajuda**

Se os problemas persistirem, execute e compartilhe:
```bash
./debug_containers.sh > debug_output.txt
```

E inclua também:
- Versão do Docker: `docker --version`
- Sistema operacional: `uname -a`
- Recursos disponíveis: `docker system df`

## 🔴 **Erro: Aplicação Encerra Durante Execução (Linha 167)**

### **Sintoma:**
```
A aplicação encerra abruptamente após a linha 167 de internal/coderunner/service.go
```

### **Causa:**
A linha 167 está no método `runPython`, especificamente no `cmd.Wait()` que aguarda o processo Python terminar. Isso indica que:
- O processo Python está crashando
- Falta de recursos (memória/CPU)
- Problemas com cgroups (Resource Management)
- Python ou dependências não estão disponíveis
- Problemas de permissões

### **Diagnóstico Rápido:**
```bash
# Script específico para investigar crashes
./debug_application_crash.sh

# Teste mínimo de execução
./test_minimal_execution.sh

# Debug completo com containers ativos
./run_image_tests.sh --debug-crash
```

### **Soluções:**

#### **1. Verificar Ambiente Python**
```bash
# Entrar no container e testar Python
docker exec -it codeactions-app-test sh

# Dentro do container:
python --version
python -c "print('Python OK')"
ls -la /app/engines/py/
```

#### **2. Desabilitar Resource Management Temporariamente**
Se o problema for relacionado a cgroups, você pode desabilitar temporariamente:
```bash
# Verificar configuração atual
docker exec codeactions-app-test env | grep RESOURCE

# Se necessário, reiniciar sem resource management
# (modificar configuração da aplicação)
```

#### **3. Verificar Logs Detalhados**
```bash
# Logs completos da aplicação
docker logs codeactions-app-test --tail 100

# Monitorar logs em tempo real
docker logs codeactions-app-test -f

# Verificar se há OOM kills
docker exec codeactions-app-test dmesg | grep -i "killed process"
```

#### **4. Testar Execução Python Manual**
```bash
# Dentro do container
docker exec -it codeactions-app-test sh

# Testar main.py diretamente
cd /app
python engines/py/main.py --help

# Criar arquivo de teste
echo 'def Run(engine): print("test")' > /tmp/test.py
python engines/py/main.py -c test123 -r run123
```

## 🔴 **Erro: Dependências Não Ficam Prontas** 