package app

import (
	"fmt"
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	authservice "sso/internal/services/auth"
	"sso/internal/storage/postgresql"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	grpcPort int,
	tokenTTL time.Duration,
) *App {
	const op = "app.New"

	storage, err := postgresql.New(cfg)
	if err != nil {
		panic(fmt.Errorf("%s: failed to init storage: %w", op, err))
	}

	authSrv := authservice.New(log, storage, storage, nil, tokenTTL)
	grpcApp := grpcapp.New(log, authSrv, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
