# Guia de Debug - Crash na Linha 167

## 🎯 **Problema:**
Aplicação encerra após linha 167 de `internal/coderunner/service.go` durante execução de código Python.

## 🔧 **Ferramentas Disponíveis:**

### **1. Debug Automático**
```bash
# Inicia containers e faz diagnóstico completo
./run_image_tests.sh --debug-crash
```

### **2. Investigação de Crash**
```bash
# Script específico para crashes (requer containers rodando)
./debug_application_crash.sh
```

### **3. Teste Mínimo**
```bash
# Testa execução Python básica
./test_minimal_execution.sh
```

### **4. Debug Manual**
```bash
# Entrar no container
docker exec -it codeactions-app-test sh

# Ver logs em tempo real
docker logs codeactions-app-test -f
```

## 🔍 **Fluxo de Investigação:**

### **Passo 1: Verificação Básica**
```bash
./test_minimal_execution.sh
```
- ✅ Se passar: problema pode ser específico de código
- ❌ Se falhar: problema é do ambiente

### **Passo 2: Diagnóstico Completo**
```bash
./run_image_tests.sh --debug-crash
```
Verifica:
- Status dos containers
- Ambiente Python
- Recursos do sistema
- Cgroups
- Permissões

### **Passo 3: Análise de Logs**
```bash
docker logs codeactions-app-test --tail 100
```
Procurar por:
- Erros de Python
- Problemas de memória
- Falhas de cgroup
- Timeouts

## 🎯 **Principais Causas e Soluções:**

### **1. Python não disponível**
**Sintoma:** `python: command not found`
**Solução:** Verificar Dockerfile, instalar Python

### **2. engines/py/main.py ausente**
**Sintoma:** `FileNotFoundError: engines/py/main.py`
**Solução:** Verificar se arquivo está na imagem Docker

### **3. Problemas de cgroups**
**Sintoma:** Erros relacionados a `/sys/fs/cgroup`
**Solução:** Desabilitar Resource Management temporariamente

### **4. Falta de recursos**
**Sintoma:** OOM kills, timeouts
**Solução:** Aumentar limites de memória/CPU

### **5. Problemas de permissões**
**Sintoma:** `Permission denied`
**Solução:** Verificar usuário que executa aplicação

## 💡 **Comandos Úteis:**

```bash
# Status dos containers
docker ps -a

# Logs da aplicação
docker logs codeactions-app-test

# Entrar no container
docker exec -it codeactions-app-test sh

# Recursos do container
docker stats codeactions-app-test

# Informações do container
docker inspect codeactions-app-test

# Teste de conectividade
./debug_containers.sh

# Limpeza completa
docker stop codeactions-app-test mongo-test redis-test
docker rm codeactions-app-test mongo-test redis-test
docker system prune -f
``` 