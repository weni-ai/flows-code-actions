package main

import (
	"log"

	"github.com/weni-ai/flows-code-actions/internal/codelib/pg"
)

func main() {
	err := pg.Example()
	if err != nil {
		log.Println(err)
	}
}
