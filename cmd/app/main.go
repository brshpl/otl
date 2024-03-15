package main

import (
	"log"

	"github.com/brshpl/otl/config"
	"github.com/brshpl/otl/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
