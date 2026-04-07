package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Emiloart/HDIP/services/issuer-api/internal/app"
	"github.com/Emiloart/HDIP/services/issuer-api/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("run service: %v", err)
	}
}
