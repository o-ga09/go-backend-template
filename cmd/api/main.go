package main

import (
	"context"
	"log"

	"github.com/o-ga09/go-backend-template/internal/server"
	"github.com/o-ga09/go-backend-template/pkg/config"
	"github.com/o-ga09/go-backend-template/pkg/logger"
)

func main() {
	ctx, err := config.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	logger.Logger(ctx)

	srv := server.New(ctx)
	if err := srv.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
