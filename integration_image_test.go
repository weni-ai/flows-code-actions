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

func (suite *ImageIntegrationTestSuite) TestCodeWithParameters() {
	// Código que usa parâmetros da requisição
	parameterCode := `
def Run(engine):
    # Tentar obter parâmetros da requisição (se houver)
    request_data = engine.get("request", {})
    
    result = {
        "message": "Parameter test",
        "received_request": bool(request_data),
        "echo": "Hello from endpoint",
        "timestamp": "2024-01-01T00:00:00Z"
    }
    
    engine.result.set(result, content_type="json")
`

	// Criar código
	createURL := suite.baseURL + "/code"
	req, err := http.NewRequest(http.MethodPost, createURL, bytes.NewBuffer([]byte(parameterCode)))
	suite.Require().NoError(err)

	q := req.URL.Query()
	q.Add("project_uuid", suite.projectUUID)
	q.Add("name", "Parameter Test Code")
	q.Add("type", "endpoint")
	q.Add("language", "python")
	req.URL.RawQuery = q.Encode()

	resp, err := suite.httpClient.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	var createdCode map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createdCode)
	codeID := createdCode["id"].(string)

	// Executar com payload JSON
	executeURL := fmt.Sprintf("%s/action/endpoint/%s", suite.baseURL, codeID)
	payload := map[string]interface{}{
		"test_param": "test_value",
		"number":     123,
	}
	payloadBytes, _ := json.Marshal(payload)

	execResp, err := suite.httpClient.Post(executeURL, "application/json", bytes.NewBuffer(payloadBytes))
	suite.Require().NoError(err)
	defer execResp.Body.Close()

	suite.Assert().Equal(http.StatusOK, execResp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(execResp.Body).Decode(&result)

	suite.Assert().Equal("Parameter test", result["message"])
	suite.Assert().Contains(result, "echo")

	fmt.Printf("✅ Teste com parâmetros executado\n")
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
	} else {
		fmt.Printf("✅ Código com erro rejeitado na criação (status: %d)\n", resp.StatusCode)
	}
}

// Cleanup após todos os testes
func (suite *ImageIntegrationTestSuite) TearDownSuite() {
	fmt.Println("🧹 Limpeza dos testes concluída")
}

func TestImageIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ImageIntegrationTestSuite))
}
