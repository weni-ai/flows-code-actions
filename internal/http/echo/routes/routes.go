package routes

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/code"
	codeRepoMongo "github.com/weni-ai/flows-code-actions/internal/code/mongodb"
	codeRepoPG "github.com/weni-ai/flows-code-actions/internal/code/pg"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	codelibRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelib/mongodb"
	codelibRepoPG "github.com/weni-ai/flows-code-actions/internal/codelib/pg"
	"github.com/weni-ai/flows-code-actions/internal/codelog"
	codelogRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelog/mongodb"
	codelogRepoS3 "github.com/weni-ai/flows-code-actions/internal/codelog/s3"
	"github.com/weni-ai/flows-code-actions/internal/coderun"
	coderunRepoMongo "github.com/weni-ai/flows-code-actions/internal/coderun/mongodb"
	coderunRepoPG "github.com/weni-ai/flows-code-actions/internal/coderun/pg"
	"github.com/weni-ai/flows-code-actions/internal/coderunner"
	s "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/http/echo/handlers"
	"github.com/weni-ai/flows-code-actions/internal/permission"
	"github.com/weni-ai/flows-code-actions/internal/secrets"
	secretsRepoPG "github.com/weni-ai/flows-code-actions/internal/secrets/pg"
	"github.com/weni-ai/flows-code-actions/internal/workerpool"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(server *s.Server) {
	healthHandler := handlers.NewHealthHandler(server)

	// Setup repositories based on database type
	var codeRepo code.Repository
	var codelibRepo codelib.Repository
	var coderunRepo coderun.Repository
	var codelogRepo codelog.Repository
	var secretsRepo secrets.Repository

	if server.Config.DB.Type == "postgres" {
		// Use PostgreSQL repositories
		pgDB := server.SQLDB
		codeRepo = codeRepoPG.NewCodeRepository(pgDB)
		codelibRepo = codelibRepoPG.NewCodeLibRepo(pgDB)
		coderunRepo = coderunRepoPG.NewCodeRunRepository(pgDB)
		secretsRepo = secretsRepoPG.NewSecretRepository(pgDB)
	} else {
		// Use MongoDB repositories (default)
		mongoDB := server.DB
		codeRepo = codeRepoMongo.NewCodeRepository(mongoDB)
		codelibRepo = codelibRepoMongo.NewCodeLibRepo(mongoDB)
		coderunRepo = coderunRepoMongo.NewCodeRunRepository(mongoDB)
		codelogRepo = codelogRepoMongo.NewCodeLogRepository(mongoDB)
	}

	codeService := code.NewCodeService(server.Config, codeRepo, codelibRepo)
	codeHandler := handlers.NewCodeHandler(codeService)

	// Setup secrets service and handler (only for PostgreSQL)
	var secretHandler *handlers.SecretHandler
	if server.Config.DB.Type == "postgres" && secretsRepo != nil {
		secretService := secrets.NewSecretService(secretsRepo)
		secretHandler = handlers.NewSecretHandler(secretService, codeService)
	}

	coderunService := coderun.NewCodeRunService(coderunRepo)
	coderunHandler := handlers.NewCodeRunHandler(coderunService)

	// Create CodeLog repository (MongoDB or S3 based on config)
	codelogRepo, err := createCodeLogRepository(server.Config, server.DB)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create codelog repository")
	}
	codelogService := codelog.NewCodeLogService(codelogRepo)
	codelogHandler := handlers.NewCodeLogHandler(codelogService, coderunService)

	server.Services.CodeLogService = codelogService
	server.Services.CodeRunService = coderunService

	coderunnerService := coderunner.NewCodeRunnerService(server.Config, coderunService, codelogService)
	pool := workerpool.NewPool(server.Config.WorkerPool.Workers, server.Config.WorkerPool.QueueSize)
	coderunnerHandler := handlers.NewCodeRunnerHandler(codeService, coderunnerService, pool)

	ratelimiter := s.NewRateLimiter(
		server.Redis,
		server.Config.RateLimiterCode.MaxRequests,
		time.Duration(server.Config.RateLimiterCode.Window)*time.Second,
	)

	log := logrus.New()
	server.Echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.WithFields(logrus.Fields{
					"URI":    v.URI,
					"status": v.Status,
				}).Info("request")
			} else {
				log.WithFields(logrus.Fields{
					"URI":    v.URI,
					"status": v.Status,
					"error":  v.Error,
				}).Error("request error")
			}
			return nil
		},
	}))

	server.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://drogasil-demo.netlify.app"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	server.Echo.GET("/", healthHandler.HealthCheck)
	server.Echo.GET("/health", healthHandler.Health)
	server.Echo.GET("/healthz", healthHandler.HealthCheck) // Simplified health check endpoint for monitoring

	server.Echo.POST("/admin/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.CreateCode, permission.WritePermission))
	server.Echo.PATCH("/admin/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.UpdateCode, permission.WritePermission))
	server.Echo.POST("/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.CreateCode, permission.WritePermission))
	server.Echo.GET("/code", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Find, permission.ReadPermission))
	server.Echo.GET("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Get, permission.ReadPermission))
	server.Echo.PATCH("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.UpdateCode, permission.WritePermission))
	server.Echo.DELETE("/code/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codeHandler.Delete, permission.WritePermission))

	server.Echo.GET("/coderun/:id", handlers.ProtectEndpointWithAuthToken(server.Config, coderunHandler.Get, permission.ReadPermission))
	server.Echo.GET("/coderun", handlers.ProtectEndpointWithAuthToken(server.Config, coderunHandler.Find, permission.ReadPermission))

	server.Echo.GET("/codelog/:id", handlers.ProtectEndpointWithAuthToken(server.Config, codelogHandler.Get, permission.ReadPermission))
	server.Echo.GET("/codelog", handlers.ProtectEndpointWithAuthToken(server.Config, codelogHandler.Find, permission.ReadPermission))

	// Secret routes (only available when using PostgreSQL)
	if secretHandler != nil {
		server.Echo.POST("/secret", handlers.ProtectEndpointWithAuthToken(server.Config, secretHandler.CreateSecret, permission.WritePermission))
		server.Echo.GET("/secret", handlers.ProtectEndpointWithAuthToken(server.Config, secretHandler.FindSecretsByCodeID, permission.ReadPermission))
		server.Echo.GET("/secret/:id", handlers.ProtectEndpointWithAuthToken(server.Config, secretHandler.GetSecret, permission.ReadPermission))
		server.Echo.PATCH("/secret/:id", handlers.ProtectEndpointWithAuthToken(server.Config, secretHandler.UpdateSecret, permission.WritePermission))
		server.Echo.DELETE("/secret/:id", handlers.ProtectEndpointWithAuthToken(server.Config, secretHandler.DeleteSecret, permission.WritePermission))
	}

	server.Echo.POST("/run/:code_id", handlers.RequireAuthToken(server.Config, coderunnerHandler.RunCode))
	server.Echo.Any("/endpoint/:code_id", coderunnerHandler.RunEndpoint)

	server.Echo.Any("/action/endpoint/:code_id", handlers.LimitByCodeIDMiddleware(coderunnerHandler.ActionEndpoint, *ratelimiter))

	server.Echo.Use(echoprometheus.NewMiddleware("codeactions"))

	server.Echo.GET("/metrics", echoprometheus.NewHandler())
}

// createCodeLogRepository creates either MongoDB or S3 repository based on configuration
func createCodeLogRepository(cfg *config.Config, db *mongo.Database) (codelog.Repository, error) {
	if cfg.S3.Enabled {
		return createS3CodeLogRepository(cfg)
	}
	return codelogRepoMongo.NewCodeLogRepository(db), nil
}

// createS3CodeLogRepository creates an S3-based repository
func createS3CodeLogRepository(cfg *config.Config) (codelog.Repository, error) {
	if cfg.S3.BucketName == "" {
		return nil, errors.New("S3 bucket name is required when S3 is enabled")
	}

	// Create AWS session
	awsConfig := &aws.Config{
		Region: aws.String(cfg.S3.Region),
	}

	// Add credentials if provided
	if cfg.S3.AccessKeyID != "" && cfg.S3.SecretAccessKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(
			cfg.S3.AccessKeyID,
			cfg.S3.SecretAccessKey,
			"",
		)
	}

	// Set custom endpoint if provided (for LocalStack or other S3-compatible services)
	if cfg.S3.Endpoint != "" {
		awsConfig.Endpoint = aws.String(cfg.S3.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true) // Required for LocalStack
		awsConfig.DisableSSL = aws.Bool(true)       // LocalStack uses HTTP by default
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AWS session")
	}

	return codelogRepoS3.NewCodeLogRepository(sess, cfg.S3.BucketName, cfg.S3.Prefix), nil
}
