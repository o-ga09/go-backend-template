package main

import (
	"context"
	"log"

	"github.com/o-ga09/go-backend-template/internal/server"
	"github.com/o-ga09/go-backend-template/pkg/config"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/logger"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

func main() {
	ctx, err := config.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	ctx = Ctx.SetRequestID(ctx, uuid.GenerateID())
	logger.Logger(ctx)

	srv := server.New(ctx)
	if err := srv.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
