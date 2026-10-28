package main

import (
	"log"

	"github.com/AlexeytsOY/information-exchange-service/internal/app"
	"github.com/AlexeytsOY/information-exchange-service/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := app.New(cfg).Run(); err != nil {
		log.Fatalf("app: %v", err)
	}
}
