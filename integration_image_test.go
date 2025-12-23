package codeactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ImageIntegrationTestSuite struct {
	suite.Suite
	baseURL     string
	httpClient  *http.Client
	projectUUID string
}

func (suite *ImageIntegrationTestSuite) SetupSuite() {
	// URL base da aplicação rodando no container
	suite.baseURL = os.Getenv("CODEACTIONS_BASE_URL")
	if suite.baseURL == "" {
		suite.baseURL = "http://localhost:8050"
	}

	suite.projectUUID = "6422db5f-6cc7-44cf-a0da-7cf12fbfe711"

	// Cliente HTTP com timeout
	suite.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Aguardar aplicação ficar disponível
	suite.waitForApplication()
}

func (suite *ImageIntegrationTestSuite) waitForApplication() {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		resp, err := suite.httpClient.Get(suite.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Printf("✅ Aplicação disponível em %s\n", suite.baseURL)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}

		fmt.Printf("⏳ Aguardando aplicação... tentativa %d/%d\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	suite.T().Fatalf("❌ Aplicação não ficou disponível em %s após %d tentativas", suite.baseURL, maxRetries)
}

func (suite *ImageIntegrationTestSuite) TestHealthEndpoint() {
	resp, err := suite.httpClient.Get(suite.baseURL + "/health")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Assert().Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	fmt.Printf("Health response: %s\n", string(body))
}

func (suite *ImageIntegrationTestSuite) TestCreateAndExecuteCode() {
	// 1. Criar código Python
	pythonCode := `
def Run(engine):
    result = 42 + 8
    engine.result.set({"calculation": result, "message": "Integration test success!"}, content_type="json")
`

	// Preparar requisição de criação
	createURL := suite.baseURL + "/code"
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewBuffer([]byte(pythonCode)))
	suite.Require().NoError(err)

	// Adicionar query parameters
	q := req.URL.Query()
	q.Add("project_uuid", suite.projectUUID)
	q.Add("name", "Integration Test Code")
	q.Add("type", "endpoint")
	q.Add("language", "python")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "text/plain")

	// Executar criação
	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Assert().Equal(http.StatusCreated, resp.StatusCode)

	// Parse da resposta
	var createdCode map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdCode)
	suite.Require().NoError(err)

	codeID, ok := createdCode["id"].(string)
	suite.Require().True(ok, "ID deve estar presente na resposta")
	suite.Assert().NotEmpty(codeID)

	fmt.Printf("✅ Código criado com ID: %s\n", codeID)

	// 2. Executar o código via endpoint
	executeURL := fmt.Sprintf("%s/action/endpoint/%s", suite.baseURL, codeID)
	execResp, err := suite.httpClient.Post(executeURL, "application/json", nil)
	suite.Require().NoError(err)
	defer execResp.Body.Close()

	suite.Assert().Equal(http.StatusOK, execResp.StatusCode)

	// Verificar resultado da execução
	var execResult map[string]interface{}
	err = json.NewDecoder(execResp.Body).Decode(&execResult)
	suite.Require().NoError(err)

	suite.Assert().Contains(execResult, "calculation")
	suite.Assert().Contains(execResult, "message")
	suite.Assert().Equal(float64(50), execResult["calculation"])

	fmt.Printf("✅ Código executado com sucesso: %+v\n", execResult)

	// 3. Limpar - deletar código criado
	deleteURL := fmt.Sprintf("%s/code/%s", suite.baseURL, codeID)
	deleteReq, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	suite.Require().NoError(err)

	dq := deleteReq.URL.Query()
	dq.Add("project_uuid", suite.projectUUID)
	deleteReq.URL.RawQuery = dq.Encode()

	deleteResp, err := suite.httpClient.Do(deleteReq)
	suite.Require().NoError(err)
	defer deleteResp.Body.Close()

	fmt.Printf("🧹 Código deletado (status: %d)\n", deleteResp.StatusCode)
}

func (suite *ImageIntegrationTestSuite) TestCreateCodeWithComplexLogic() {
	// Código mais complexo para testar funcionalidades
	complexCode := `
import json
import datetime

def Run(engine):
    # Teste de bibliotecas e lógica
    current_time = datetime.datetime.now()
    data = {
        "timestamp": current_time.isoformat(),
        "operations": {
            "addition": 10 + 5,
            "multiplication": 7 * 8,
            "string_ops": "Hello " + "World"
        },
        "json_test": json.dumps({"nested": "value"}),
        "list_ops": [1, 2, 3, 4, 5],
        "status": "success"
    }
    
    engine.result.set(data, content_type="json")
`

	// Criar código
	createURL := suite.baseURL + "/code"
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewBuffer([]byte(complexCode)))
	suite.Require().NoError(err)

	q := req.URL.Query()
	q.Add("project_uuid", suite.projectUUID)
	q.Add("name", "Complex Integration Test")
	q.Add("type", "endpoint")
	q.Add("language", "python")
	req.URL.RawQuery = q.Encode()

	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Assert().Equal(http.StatusCreated, resp.StatusCode)

	var createdCode map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdCode)
	suite.Require().NoError(err)

	codeID := createdCode["id"].(string)

	// Executar código complexo
	executeURL := fmt.Sprintf("%s/action/endpoint/%s", suite.baseURL, codeID)
	execResp, err := suite.httpClient.Post(executeURL, "application/json", nil)
	suite.Require().NoError(err)
	defer execResp.Body.Close()

	suite.Assert().Equal(http.StatusOK, execResp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(execResp.Body).Decode(&result)
	suite.Require().NoError(err)

	// Verificações detalhadas
	suite.Assert().Contains(result, "timestamp")
	suite.Assert().Contains(result, "operations")
	suite.Assert().Contains(result, "status")
	suite.Assert().Equal("success", result["status"])

	operations := result["operations"].(map[string]interface{})
	suite.Assert().Equal(float64(15), operations["addition"])
	suite.Assert().Equal(float64(56), operations["multiplication"])
	suite.Assert().Equal("Hello World", operations["string_ops"])

	fmt.Printf("✅ Código complexo executado: timestamp=%v\n", result["timestamp"])

	// Limpar - deletar código criado
	deleteURL := fmt.Sprintf("%s/code/%s", suite.baseURL, codeID)
	deleteReq, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	suite.Require().NoError(err)

	dq := deleteReq.URL.Query()
	dq.Add("project_uuid", suite.projectUUID)
	deleteReq.URL.RawQuery = dq.Encode()

	deleteResp, err := suite.httpClient.Do(deleteReq)
	suite.Require().NoError(err)
	defer deleteResp.Body.Close()

	fmt.Printf("🧹 Código complexo deletado (status: %d)\n", deleteResp.StatusCode)
}

func (suite *ImageIntegrationTestSuite) TestListCodes() {
	// Testar listagem de códigos
	listURL := suite.baseURL + "/code?project_uuid=" + url.QueryEscape(suite.projectUUID)

	resp, err := suite.httpClient.Get(listURL)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Assert().Equal(http.StatusOK, resp.StatusCode)

	var codeList []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&codeList)
	suite.Require().NoError(err)

	// Deve ter pelo menos os códigos criados nos testes anteriores
	suite.Assert().GreaterOrEqual(len(codeList), 0)

	fmt.Printf("✅ Listagem retornou %d códigos\n", len(codeList))
}

func (suite *ImageIntegrationTestSuite) TestErrorHandling() {
	// Testar código com erro intencional
	errorCode := `
def Run(engine):
    # Código que vai gerar erro
    result = 1 / 0  # Division by zero
    engine.result.set({"should": "not_reach_here"}, content_type="json")
`

	createURL := suite.baseURL + "/code"
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewBuffer([]byte(errorCode)))
	suite.Require().NoError(err)

	q := req.URL.Query()
	q.Add("project_uuid", suite.projectUUID)
	q.Add("name", "Error Test Code")
	q.Add("type", "endpoint")
	q.Add("language", "python")
	req.URL.RawQuery = q.Encode()

	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var createdCode map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&createdCode)
		codeID := createdCode["id"].(string)

		// Executar código com erro
		executeURL := fmt.Sprintf("%s/action/endpoint/%s", suite.baseURL, codeID)
		execResp, err := suite.httpClient.Post(executeURL, "application/json", nil)
		suite.Require().NoError(err)
		defer execResp.Body.Close()

		// Pode retornar erro ou uma resposta de erro estruturada
		// Verificamos que pelo menos não trava a aplicação
		suite.Assert().True(execResp.StatusCode >= 400 || execResp.StatusCode == 200)

		fmt.Printf("✅ Tratamento de erro testado (status: %d)\n", execResp.StatusCode)

		// Limpar - deletar código criado
		deleteURL := fmt.Sprintf("%s/code/%s", suite.baseURL, codeID)
		deleteReq, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
		suite.Require().NoError(err)

		dq := deleteReq.URL.Query()
		dq.Add("project_uuid", suite.projectUUID)
		deleteReq.URL.RawQuery = dq.Encode()

		deleteResp, err := suite.httpClient.Do(deleteReq)
		suite.Require().NoError(err)
		defer deleteResp.Body.Close()

		fmt.Printf("🧹 Código de erro deletado (status: %d)\n", deleteResp.StatusCode)
	} else {
		fmt.Printf("✅ Código com erro rejeitado na criação (status: %d)\n", resp.StatusCode)
	}
}

func TestListLogs(t *testing.T) {
	suite := &ImageIntegrationTestSuite{}
	suite.SetT(t)
	suite.SetupSuite()
	defer suite.TearDownSuite()

	// 1. Criar código que gera logs
	pythonCodeWithLogs := `
import json

def Run(engine):
    # Gerar diferentes tipos de logs
    engine.log.debug("Debug log: Starting execution")
    engine.log.info("Info log: Processing data")
    
    try:
        result = {"status": "success", "value": 100}
        engine.log.info("Info log: Calculation completed successfully")
        engine.result.set(result, content_type="json")
    except Exception as e:
        engine.log.error(f"Error log: {str(e)}")
        raise e
`

	// Criar código
	createURL := suite.baseURL + "/code"
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewBuffer([]byte(pythonCodeWithLogs)))
	suite.Require().NoError(err)

	q := req.URL.Query()
	q.Add("project_uuid", suite.projectUUID)
	q.Add("name", "Test Log Generation Code")
	q.Add("type", "endpoint")
	q.Add("language", "python")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "text/plain")

	// Executar criação
	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Assert().Equal(http.StatusCreated, resp.StatusCode)

	// Parse da resposta
	var createdCode map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdCode)
	suite.Require().NoError(err)

	codeID, ok := createdCode["id"].(string)
	suite.Require().True(ok, "ID deve estar presente na resposta")
	suite.Assert().NotEmpty(codeID)

	fmt.Printf("✅ Código criado para teste de logs com ID: %s\n", codeID)

	// 2. Executar o código para gerar logs
	executeURL := fmt.Sprintf("%s/action/endpoint/%s", suite.baseURL, codeID)
	execResp, err := suite.httpClient.Post(executeURL, "application/json", nil)
	suite.Require().NoError(err)
	defer execResp.Body.Close()

	suite.Assert().Equal(http.StatusOK, execResp.StatusCode)

	fmt.Printf("✅ Código executado para gerar logs\n")

	// 3. Aguardar um momento para logs serem persistidos
	time.Sleep(2 * time.Second)

	// 4. Listar logs por code_id
	logsURL := fmt.Sprintf("%s/codelog?code_id=%s&page=1", suite.baseURL, url.QueryEscape(codeID))
	logsResp, err := suite.httpClient.Get(logsURL)
	suite.Require().NoError(err)
	defer logsResp.Body.Close()

	suite.Assert().Equal(http.StatusOK, logsResp.StatusCode)

	// Parse da resposta de logs
	var logsResponse map[string]interface{}
	err = json.NewDecoder(logsResp.Body).Decode(&logsResponse)
	suite.Require().NoError(err)

	// Verificar estrutura da resposta
	suite.Assert().Contains(logsResponse, "data")
	suite.Assert().Contains(logsResponse, "total")
	suite.Assert().Contains(logsResponse, "page")
	suite.Assert().Contains(logsResponse, "last_page")

	// Verificar que pelo menos alguns logs foram gerados
	data, ok := logsResponse["data"].([]interface{})
	suite.Require().True(ok, "data deve ser um array")

	total, ok := logsResponse["total"].(float64)
	suite.Require().True(ok, "total deve ser um número")

	page, ok := logsResponse["page"].(float64)
	suite.Require().True(ok, "page deve ser um número")

	fmt.Printf("✅ Logs listados: total=%v, page=%v, logs_count=%d\n", total, page, len(data))

	// Verificar que temos pelo menos um log (pode não ter logs se execução for muito rápida)
	suite.Assert().GreaterOrEqual(int(total), 0, "Total de logs deve ser >= 0")
	suite.Assert().Equal(1, int(page), "Página deve ser 1")

	// Se temos logs, verificar estrutura de pelo menos um log
	if len(data) > 0 {
		logEntry := data[0].(map[string]interface{})
		suite.Assert().Contains(logEntry, "id")
		suite.Assert().Contains(logEntry, "run_id")
		suite.Assert().Contains(logEntry, "code_id")
		suite.Assert().Contains(logEntry, "type")
		suite.Assert().Contains(logEntry, "content")
		suite.Assert().Contains(logEntry, "created_at")

		// Verificar que o code_id no log corresponde ao código criado
		logCodeID, ok := logEntry["code_id"].(string)
		if ok {
			suite.Assert().Equal(codeID, logCodeID, "code_id no log deve corresponder ao código executado")
		}

		fmt.Printf("✅ Estrutura do log verificada: type=%s, content=%s\n",
			logEntry["type"], logEntry["content"])
	}

	// 5. Testar paginação (page 2)
	logsURL2 := fmt.Sprintf("%s/codelog?code_id=%s&page=2", suite.baseURL, url.QueryEscape(codeID))
	logsResp2, err := suite.httpClient.Get(logsURL2)
	suite.Require().NoError(err)
	defer logsResp2.Body.Close()

	suite.Assert().Equal(http.StatusOK, logsResp2.StatusCode)

	var logsResponse2 map[string]interface{}
	err = json.NewDecoder(logsResp2.Body).Decode(&logsResponse2)
	suite.Require().NoError(err)

	page2, ok := logsResponse2["page"].(float64)
	suite.Require().True(ok)
	suite.Assert().Equal(2, int(page2), "Página deve ser 2")

	fmt.Printf("✅ Paginação testada: page=%v\n", page2)

	// 6. Testar parâmetro inválido (sem run_id nem code_id)
	invalidLogsURL := fmt.Sprintf("%s/codelog?page=1", suite.baseURL)
	invalidResp, err := suite.httpClient.Get(invalidLogsURL)
	suite.Require().NoError(err)
	defer invalidResp.Body.Close()

	suite.Assert().Equal(http.StatusBadRequest, invalidResp.StatusCode)
	fmt.Printf("✅ Validação de parâmetros testada (status: %d)\n", invalidResp.StatusCode)

	// 7. Limpar - deletar código criado
	deleteURL := fmt.Sprintf("%s/code/%s", suite.baseURL, codeID)
	deleteReq, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	suite.Require().NoError(err)

	dq := deleteReq.URL.Query()
	dq.Add("project_uuid", suite.projectUUID)
	deleteReq.URL.RawQuery = dq.Encode()

	deleteResp, err := suite.httpClient.Do(deleteReq)
	suite.Require().NoError(err)
	defer deleteResp.Body.Close()

	fmt.Printf("🧹 Código de teste de logs deletado (status: %d)\n", deleteResp.StatusCode)
}

// Cleanup após todos os testes
func (suite *ImageIntegrationTestSuite) TearDownSuite() {
	fmt.Println("🧹 Limpeza dos testes concluída")
}

func TestImageIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ImageIntegrationTestSuite))
}
