package main

import (
	app "github.com/weni-ai/code-actions"
	"github.com/weni-ai/code-actions/config"
)

func main() {
	cfg := config.NewConfig()

	app.Start(cfg)
}
