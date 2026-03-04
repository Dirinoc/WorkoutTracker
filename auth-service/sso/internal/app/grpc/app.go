package grpcapp

import (
	"fmt"

	"net"
	authrpc "sso/internal/grpc/auth"

	"google.golang.org/grpc"
)

type App struct {
	gRPCServer *grpc.Server
	port       int
}

func New(port int) *App {
	gRPCServer := grpc.NewServer()

	authrpc.Register(gRPCServer)

	return &App{
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", err)
	}

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

//TODO: Graceful shutdown
