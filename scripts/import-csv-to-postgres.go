package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Script para importar dados CSV exportados do MongoDB para o PostgreSQL
// Uso: go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231201_120000

var (
	csvDir     = flag.String("dir", "", "Diretório contendo os arquivos CSV exportados do MongoDB")
	pgURI      = flag.String("pg-uri", "", "URI de conexão do PostgreSQL (padrão: variável de ambiente)")
	dryRun     = flag.Bool("dry-run", false, "Executar sem fazer alterações no banco")
	batchSize  = flag.Int("batch-size", 1000, "Tamanho do lote para inserções em batch")
	skipErrors = flag.Bool("skip-errors", false, "Continuar mesmo se houver erros em alguns registros")
)

type ImportStats struct {
	Total    int
	Success  int
	Skipped  int
	Failed   int
	Duration time.Duration
}

func main() {
	flag.Parse()

	if *csvDir == "" {
		fmt.Println("Erro: é necessário especificar o diretório com -dir")
		fmt.Println("Uso: go run scripts/import-csv-to-postgres.go -dir=./mongo_exports/20231201_120000")
		os.Exit(1)
	}

	// Obter URI do PostgreSQL
	dbURI := *pgURI
	if dbURI == "" {
		dbURI = os.Getenv("FLOWS_CODE_ACTIONS_DB_URI")
		if dbURI == "" {
			dbURI = "postgres://test:test@localhost:5432/codeactions?sslmode=disable"
		}
	}

	fmt.Println("============================================")
	fmt.Println("Importação de CSV para PostgreSQL")
	fmt.Println("============================================")
	fmt.Printf("Diretório CSV: %s\n", *csvDir)
	fmt.Printf("PostgreSQL URI: %s\n", maskPassword(dbURI))
	if *dryRun {
		fmt.Println("MODO DRY-RUN: Nenhuma alteração será feita")
	}
	fmt.Println("============================================")
	fmt.Println()

	// Verificar se o diretório existe
	if _, err := os.Stat(*csvDir); os.IsNotExist(err) {
		fmt.Printf("Erro: diretório não encontrado: %s\n", *csvDir)
		os.Exit(1)
	}

	ctx := context.Background()

	// Conectar ao PostgreSQL
	var db *sql.DB
	var err error

	if !*dryRun {
		db, err = sql.Open("postgres", dbURI)
		if err != nil {
			fmt.Printf("Erro ao conectar no PostgreSQL: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Testar conexão
		if err := db.PingContext(ctx); err != nil {
			fmt.Printf("Erro ao fazer ping no PostgreSQL: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Conectado ao PostgreSQL com sucesso")
		fmt.Println()
	}

	// Importar cada collection na ordem correta (respeitando dependências)
	collections := []string{
		"code",
		"codelib",
		"coderun",
		"projects",
		"user_permissions",
	}

	totalStats := make(map[string]*ImportStats)

	for _, collection := range collections {
		csvFile := filepath.Join(*csvDir, collection+".csv")

		// Verificar se arquivo existe
		fileInfo, err := os.Stat(csvFile)
		if os.IsNotExist(err) {
			fmt.Printf("⚠ Arquivo não encontrado: %s (pulando)\n\n", csvFile)
			continue
		}

		// Verificar se arquivo está vazio
		if fileInfo.Size() == 0 {
			fmt.Printf("⚠ Arquivo vazio: %s (pulando)\n\n", csvFile)
			continue
		}

		fmt.Printf("Importando: %s\n", collection)
		stats, err := importCollection(ctx, db, collection, csvFile)
		if err != nil {
			// Tratar erros específicos que devem pular para próxima collection
			errMsg := err.Error()
			if strings.Contains(errMsg, "EOF") ||
				strings.Contains(errMsg, "erro ao ler cabeçalho") ||
				strings.Contains(errMsg, "arquivo vazio") {
				fmt.Printf("⚠ Arquivo %s vazio ou inválido: %v (pulando)\n\n", collection, err)
				continue
			}

			if *skipErrors {
				fmt.Printf("✗ Erro ao importar %s: %v (continuando...)\n\n", collection, err)
				continue
			} else {
				fmt.Printf("✗ Erro ao importar %s: %v\n", collection, err)
				os.Exit(1)
			}
		}

		totalStats[collection] = stats
		fmt.Printf("✓ %s importado: %d total, %d sucesso, %d falhas, %d pulados (%.2fs)\n\n",
			collection, stats.Total, stats.Success, stats.Failed, stats.Skipped,
			stats.Duration.Seconds())
	}

	// Resumo final
	fmt.Println("============================================")
	fmt.Println("Resumo da Importação")
	fmt.Println("============================================")

	totalImported := 0
	totalRecords := 0

	for collection, stats := range totalStats {
		fmt.Printf("%s: %d/%d registros importados\n", collection, stats.Success, stats.Total)
		totalImported += stats.Success
		totalRecords += stats.Total
	}

	if len(totalStats) == 0 {
		fmt.Println("⚠ Nenhuma collection foi importada")
	} else {
		fmt.Printf("\nTotal: %d/%d registros importados\n", totalImported, totalRecords)
	}

	fmt.Println("============================================")
}

func importCollection(ctx context.Context, db *sql.DB, collection, csvFile string) (*ImportStats, error) {
	stats := &ImportStats{}
	startTime := time.Now()

	file, err := os.Open(csvFile)
	if err != nil {
		return stats, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Ler cabeçalho
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return stats, fmt.Errorf("arquivo vazio (sem cabeçalho)")
		}
		return stats, fmt.Errorf("erro ao ler cabeçalho: %w", err)
	}

	// Verificar se tem pelo menos um campo no cabeçalho
	if len(headers) == 0 {
		return stats, fmt.Errorf("arquivo vazio (cabeçalho sem campos)")
	}

	// Processar registros
	batch := [][]string{}
	lineNum := 1 // Começar em 1 (cabeçalho)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if *skipErrors {
				stats.Failed++
				fmt.Printf("  ⚠ Linha %d: erro ao ler CSV: %v\n", lineNum+1, err)
				continue
			}
			return stats, fmt.Errorf("erro ao ler linha %d: %w", lineNum+1, err)
		}

		lineNum++
		stats.Total++

		// Validar número de campos
		if len(record) != len(headers) {
			if *skipErrors {
				stats.Failed++
				fmt.Printf("  ⚠ Linha %d: número de campos inválido (esperado %d, obtido %d)\n",
					lineNum, len(headers), len(record))
				continue
			}
			return stats, fmt.Errorf("linha %d: número de campos inválido", lineNum)
		}

		batch = append(batch, record)

		// Processar em lotes
		if len(batch) >= *batchSize {
			if err := processBatch(ctx, db, collection, headers, batch); err != nil {
				if *skipErrors {
					stats.Failed += len(batch)
					fmt.Printf("  ⚠ Erro ao processar lote: %v\n", err)
				} else {
					return stats, err
				}
			} else {
				stats.Success += len(batch)
			}
			batch = [][]string{}
		}
	}

	// Processar registros restantes
	if len(batch) > 0 {
		if err := processBatch(ctx, db, collection, headers, batch); err != nil {
			if *skipErrors {
				stats.Failed += len(batch)
				fmt.Printf("  ⚠ Erro ao processar último lote: %v\n", err)
			} else {
				return stats, err
			}
		} else {
			stats.Success += len(batch)
		}
	}

	stats.Duration = time.Since(startTime)
	return stats, nil
}

func processBatch(ctx context.Context, db *sql.DB, collection string, headers []string, batch [][]string) error {
	if *dryRun {
		return nil
	}

	switch collection {
	case "code":
		return importCodeBatch(ctx, db, headers, batch)
	case "codelib":
		return importCodelibBatch(ctx, db, headers, batch)
	case "coderun":
		return importCoderunBatch(ctx, db, headers, batch)
	case "projects":
		return importProjectsBatch(ctx, db, headers, batch)
	case "user_permissions":
		return importUserPermissionsBatch(ctx, db, headers, batch)
	default:
		return fmt.Errorf("collection desconhecida: %s", collection)
	}
}

func importCodeBatch(ctx context.Context, db *sql.DB, headers []string, batch [][]string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range batch {
		data := makeMap(headers, record)

		// Converter _id do MongoDB para mongo_object_id
		mongoID := data["_id"]
		if mongoID == "" {
			continue
		}

		// Parse dos campos
		timeout := 60
		if t := data["timeout"]; t != "" {
			fmt.Sscanf(t, "%d", &timeout)
		}

		createdAt, _ := parseDate(data["created_at"])
		updatedAt, _ := parseDate(data["updated_at"])

		query := `
			INSERT INTO codes (mongo_object_id, name, type, source, language, url, project_uuid, timeout, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (mongo_object_id) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type,
				source = EXCLUDED.source,
				language = EXCLUDED.language,
				url = EXCLUDED.url,
				project_uuid = EXCLUDED.project_uuid,
				timeout = EXCLUDED.timeout,
				updated_at = EXCLUDED.updated_at
		`

		_, err := tx.ExecContext(ctx, query,
			mongoID,
			data["name"],
			data["type"],
			data["source"],
			data["language"],
			data["url"],
			data["project_uuid"],
			timeout,
			createdAt,
			updatedAt,
		)

		if err != nil {
			return fmt.Errorf("erro ao inserir code %s: %w", mongoID, err)
		}
	}

	return tx.Commit()
}

func importCodelibBatch(ctx context.Context, db *sql.DB, headers []string, batch [][]string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range batch {
		data := makeMap(headers, record)

		mongoID := data["_id"]
		if mongoID == "" {
			continue
		}

		createdAt, _ := parseDate(data["created_at"])
		updatedAt, _ := parseDate(data["updated_at"])

		query := `
			INSERT INTO codelibs (mongo_object_id, name, language, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (mongo_object_id) DO UPDATE SET
				name = EXCLUDED.name,
				language = EXCLUDED.language,
				updated_at = EXCLUDED.updated_at
		`

		_, err := tx.ExecContext(ctx, query,
			mongoID,
			data["name"],
			data["language"],
			createdAt,
			updatedAt,
		)

		if err != nil {
			return fmt.Errorf("erro ao inserir codelib %s: %w", mongoID, err)
		}
	}

	return tx.Commit()
}

func importCoderunBatch(ctx context.Context, db *sql.DB, headers []string, batch [][]string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range batch {
		data := makeMap(headers, record)

		mongoID := data["_id"]
		if mongoID == "" {
			continue
		}

		// Buscar code_id a partir do code_mongo_id
		var codeID *string
		if codeMongoID := data["code_id"]; codeMongoID != "" {
			var id string
			err := tx.QueryRowContext(ctx, "SELECT id FROM codes WHERE mongo_object_id = $1", codeMongoID).Scan(&id)
			if err == nil {
				codeID = &id
			} else if err != sql.ErrNoRows {
				return fmt.Errorf("erro ao buscar code_id: %w", err)
			}
		}

		createdAt, _ := parseDate(data["created_at"])
		updatedAt, _ := parseDate(data["updated_at"])

		// Parse JSON fields
		extra := parseJSON(data["extra"])
		params := parseJSON(data["params"])
		headersJSON := parseJSON(data["headers"])

		query := `
			INSERT INTO coderuns (mongo_object_id, code_id, code_mongo_id, status, result, extra, params, body, headers, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (mongo_object_id) DO UPDATE SET
				status = EXCLUDED.status,
				result = EXCLUDED.result,
				extra = EXCLUDED.extra,
				params = EXCLUDED.params,
				body = EXCLUDED.body,
				headers = EXCLUDED.headers,
				updated_at = EXCLUDED.updated_at
		`

		extraJSON, _ := json.Marshal(extra)
		paramsJSON, _ := json.Marshal(params)
		headersJSONBytes, _ := json.Marshal(headersJSON)

		_, err := tx.ExecContext(ctx, query,
			mongoID,
			codeID,
			data["code_id"],
			data["status"],
			data["result"],
			string(extraJSON),
			string(paramsJSON),
			data["body"],
			string(headersJSONBytes),
			createdAt,
			updatedAt,
		)

		if err != nil {
			return fmt.Errorf("erro ao inserir coderun %s: %w", mongoID, err)
		}
	}

	return tx.Commit()
}

func importProjectsBatch(ctx context.Context, db *sql.DB, headers []string, batch [][]string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range batch {
		data := makeMap(headers, record)

		mongoID := data["_id"]
		if mongoID == "" {
			continue
		}

		createdAt, _ := parseDate(data["created_at"])
		updatedAt, _ := parseDate(data["updated_at"])

		query := `
			INSERT INTO projects (mongo_object_id, uuid, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (mongo_object_id) DO UPDATE SET
				uuid = EXCLUDED.uuid,
				name = EXCLUDED.name,
				updated_at = EXCLUDED.updated_at
		`

		_, err := tx.ExecContext(ctx, query,
			mongoID,
			data["uuid"],
			data["name"],
			createdAt,
			updatedAt,
		)

		if err != nil {
			return fmt.Errorf("erro ao inserir project %s: %w", mongoID, err)
		}
	}

	return tx.Commit()
}

func importUserPermissionsBatch(ctx context.Context, db *sql.DB, headers []string, batch [][]string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range batch {
		data := makeMap(headers, record)

		mongoID := data["_id"]
		if mongoID == "" {
			continue
		}

		createdAt, _ := parseDate(data["created_at"])
		updatedAt, _ := parseDate(data["updated_at"])

		query := `
			INSERT INTO user_permissions (mongo_object_id, user_email, project_uuid, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (mongo_object_id) DO UPDATE SET
				user_email = EXCLUDED.user_email,
				project_uuid = EXCLUDED.project_uuid,
				role = EXCLUDED.role,
				updated_at = EXCLUDED.updated_at
		`

		_, err := tx.ExecContext(ctx, query,
			mongoID,
			data["user_email"],
			data["project_uuid"],
			data["role"],
			createdAt,
			updatedAt,
		)

		if err != nil {
			return fmt.Errorf("erro ao inserir user_permission %s: %w", mongoID, err)
		}
	}

	return tx.Commit()
}

// Funções auxiliares

func makeMap(headers, values []string) map[string]string {
	m := make(map[string]string)
	for i, header := range headers {
		if i < len(values) {
			m[header] = values[i]
		}
	}
	return m
}

func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}

	// Tentar vários formatos de data
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Now(), fmt.Errorf("formato de data não reconhecido: %s", dateStr)
}

func parseJSON(jsonStr string) interface{} {
	if jsonStr == "" || jsonStr == "{}" {
		return map[string]interface{}{}
	}

	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return map[string]interface{}{}
	}
	return result
}

func maskPassword(uri string) string {
	if strings.Contains(uri, "@") {
		parts := strings.Split(uri, "@")
		if len(parts) >= 2 {
			credentials := strings.Split(parts[0], "://")
			if len(credentials) >= 2 {
				userPass := strings.Split(credentials[1], ":")
				if len(userPass) >= 2 {
					return strings.Replace(uri, userPass[1], "****", 1)
				}
			}
		}
	}
	return uri
}
