package codeactions

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/weni-ai/code-actions/config"
	server "github.com/weni-ai/code-actions/internal/http/echo"
	"github.com/weni-ai/code-actions/internal/http/echo/routes"
	"github.com/weni-ai/code-actions/pkg/db"
)

func Start(cfg *config.Config) {
	codeactions := server.NewServer(cfg)

	db, err := db.GetMongoDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	codeactions.DB = db

	routes.Setup(codeactions)

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
