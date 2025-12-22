package codeactions

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	lockerRedis "github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	codelibRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelib/mongodb"
	codelibRepoPG "github.com/weni-ai/flows-code-actions/internal/codelib/pg"
	"github.com/weni-ai/flows-code-actions/internal/db"
	"github.com/weni-ai/flows-code-actions/internal/db/postgre"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven/rabbitmq"
	server "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/http/echo/routes"
	"github.com/weni-ai/flows-code-actions/internal/permission"
	permRepoMongo "github.com/weni-ai/flows-code-actions/internal/permission/mongodb"
	permRepoPG "github.com/weni-ai/flows-code-actions/internal/permission/pg"
	"github.com/weni-ai/flows-code-actions/internal/project"
	projRepoMongo "github.com/weni-ai/flows-code-actions/internal/project/mongodb"
	projRepoPG "github.com/weni-ai/flows-code-actions/internal/project/pg"
)

func Start(cfg *config.Config) {
	codeactions := server.NewServer(cfg)

	// Setup database based on type
	if cfg.DB.Type == "postgres" {
		log.Info("Using PostgreSQL database")
		sqlDB, err := postgre.GetPostgreDatabase(cfg)
		if err != nil {
			log.WithError(err).Fatal("Failed to connect to PostgreSQL")
		}
		codeactions.SQLDB = sqlDB
	} else {
		log.Info("Using MongoDB database")
		mongoDB, err := db.GetMongoDatabase(cfg)
		if err != nil {
			log.WithError(err).Fatal("Failed to connect to MongoDB")
		}
		codeactions.DB = mongoDB
	}

	redisURL, _ := url.Parse(cfg.Redis)
	rdb, err := strconv.Atoi(strings.TrimLeft(redisURL.Path, "/"))
	if err != nil {
		log.Fatal(err)
	}
	rpass, _ := redisURL.User.Password()
	RedisClient := redis.NewClient(&redis.Options{
		Addr:     redisURL.Host,
		DB:       rdb,
		Password: rpass,
	})
	pong, err := RedisClient.Ping(context.TODO()).Result()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("Pong:", pong)
	}
	defer RedisClient.Close()

	codeactions.Redis = RedisClient

	routes.Setup(codeactions)

	redisOpt, err := lockerRedis.ParseURL(cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	lockerRedisClient := lockerRedis.NewClient(redisOpt)
	locker := redislock.New(lockerRedisClient)

	codeactions.Locker = locker

	go func() {
		err := codeactions.Start(cfg.HTTP.Port)
		if err != nil {
			log.WithError(err).Fatal(err)
		}
	}()

	if cfg.EDA.RabbitmqURL != "" {
		eda := rabbitmq.NewEDA(cfg.EDA.RabbitmqURL)

		var permissionService permission.UserPermissionUseCase
		var projectService project.UseCase

		if cfg.DB.Type == "postgres" {
			// Use PostgreSQL repositories
			permissionService = permission.NewUserPermissionService(
				permRepoPG.NewUserRepository(codeactions.SQLDB),
			)
			projectService = project.NewProjectService(
				projRepoPG.NewProjectRepository(codeactions.SQLDB),
			)
		} else {
			// Use MongoDB repositories
			permissionService = permission.NewUserPermissionService(
				permRepoMongo.NewUserRepository(codeactions.DB),
			)
			projectService = project.NewProjectService(
				projRepoMongo.NewProjectRepository(codeactions.DB),
			)
		}

		server.Permission = server.NewEchoPermissionHandler(permissionService)

		projectConsumer := project.NewProjectConsumer(
			projectService,
			permissionService,
			cfg.EDA.ProjectExchangeName,
			cfg.EDA.ProjectQueueName,
		)
		permissionConsumer := permission.NewPermissionConsumer(
			permissionService,
			cfg.EDA.PermissionExchangeName,
			cfg.EDA.PermissionQueueName,
		)
		eda.AddConsumer(projectConsumer)
		eda.AddConsumer(permissionConsumer)

		if err := eda.StartConsumers(); err != nil {
			log.WithError(err)
		}
	}

	codeactions.StartCodeLogCleaner(context.TODO(), cfg)
	codeactions.StartCodeRunCleaner(context.TODO(), cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Signal to quit received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := codeactions.Stop(ctx); err != nil {
		log.Fatalf("Stop failed: %v\n", err)
	}
}

func SetupLibs(s *server.Server) error {
	var codelibRepo codelib.Repository

	if s.Config.DB.Type == "postgres" {
		codelibRepo = codelibRepoPG.NewCodeLibRepo(s.SQLDB)
	} else {
		codelibRepo = codelibRepoMongo.NewCodeLibRepo(s.DB)
	}

	codelibService := codelib.NewCodeLibService(codelibRepo)

	{ // setup py libs
		langPy := codelib.LanguageType(codelib.TypePy)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		currentLibs, err := codelibService.List(ctx, &langPy)
		if err != nil {
			return err
		}
		libs := []string{}
		for _, lib := range currentLibs {
			libs = append(libs, lib.Name)
		}
		if err := codelib.InstallPythonLibs(libs); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
