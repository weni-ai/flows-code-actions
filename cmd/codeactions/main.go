package main

import (
	app "github.com/weni-ai/flows-code-actions"
	"github.com/weni-ai/flows-code-actions/config"
)

func main() {
	cfg := config.NewConfig()

	app.Start(cfg)
}
