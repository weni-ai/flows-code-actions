package codeactions

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/weni-ai/flows-code-actions/config"
	"github.com/weni-ai/flows-code-actions/internal/codelib"
	codelibRepoMongo "github.com/weni-ai/flows-code-actions/internal/codelib/mongodb"
	server "github.com/weni-ai/flows-code-actions/internal/http/echo"
	"github.com/weni-ai/flows-code-actions/internal/http/echo/routes"
	"github.com/weni-ai/flows-code-actions/pkg/db"
)

func Start(cfg *config.Config) {
	codeactions := server.NewServer(cfg)

	db, err := db.GetMongoDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	codeactions.DB = db

	routes.Setup(codeactions)

	if err := SetupLibs(codeactions); err != nil {
		log.Fatal(err)
	}

	go func() {
		err := codeactions.Start(cfg.HTTP.Port)
		if err != nil {
			log.Fatal(err)
		}
	}()

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
