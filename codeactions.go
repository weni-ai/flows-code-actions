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
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	codelibRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelib/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/db"
	"github.com/weni-ai/flows-code-actions/internal/eventdriven/rabbitmq"
	server "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/http/echo/routes"
	"github.com/weni-ai/flows-code-actions/internal/permission"
	permRepoMongo "github.com/weni-ai/flows-code-actions/internal/permission/mongodb"
	"github.com/weni-ai/flows-code-actions/internal/project"
	projRepoMongo "github.com/weni-ai/flows-code-actions/internal/project/mongodb"
)

func Start(cfg *config.Config) {
	codeactions := server.NewServer(cfg)

	db, err := db.GetMongoDatabase(cfg)
	if err != nil {
		log.WithError(err).Fatal(err)
	}

	codeactions.DB = db

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

	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal(err)
	}
	redisClient := redis.NewClient(redisOpt)
	locker := redislock.New(redisClient)

	codeactions.Locker = locker

	if err := SetupLibs(codeactions); err != nil {
		log.WithError(err).Fatal(err)
	}

	go func() {
		err := codeactions.Start(cfg.HTTP.Port)
		if err != nil {
			log.WithError(err).Fatal(err)
		}
	}()

	if cfg.EDA.RabbitmqURL != "" {
		eda := rabbitmq.NewEDA(cfg.EDA.RabbitmqURL)

		permissionService := permission.NewUserPermissionService(
			permRepoMongo.NewUserRepository(db),
		)

		server.Permission = server.NewEchoPermissionHandler(permissionService)

		projectConsumer := project.NewProjectConsumer(
			project.NewProjectService(
				projRepoMongo.NewProjectRepository(db),
			),
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
	codelibRepo := codelibRepoMongo.NewCodeLibRepo(s.DB)
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
