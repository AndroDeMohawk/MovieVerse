package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AndroDeMohawk/MovieVerse/internal/client/auth"
	"github.com/AndroDeMohawk/MovieVerse/internal/config"
	"github.com/AndroDeMohawk/MovieVerse/internal/repository/postgres"
	"github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/interceptor"
	moviegrpc "github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/movie"

	"google.golang.org/grpc"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting movie-service", slog.String("env", cfg.Env))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()
	log.Info("connected to postgresql successfully")

	//Подключение gRPC-клиента к SSO сервису
	authClient, err := auth.New(ctx, cfg.AuthService.Address, cfg.AuthService.AppID, log)
	if err != nil {
		log.Error("failed to connect to auth service", slog.String("error", err.Error()))
		os.Exit(1)
	}
	_ = authClient //Для Usecase и Transpost
	log.Info("connected to auth sso service successfully", slog.String("addr", cfg.AuthService.Address))

	//Создание gRPC сервера с JWT Interceptor
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.AuthUnaryInterceptor(cfg.AuthService.AppSecret, cfg.AuthService.AppID),
		),
	)

	moviegrpc.Register(gRPCServer)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to listen port", slog.Int("port", cfg.GRPC.Port), slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		log.Info("gRPC server started", slog.String("addr", l.Addr().String()))
		if err := gRPCServer.Serve(l); err != nil {
			log.Error("gRPC server failed", slog.String("error", err.Error()))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))

	gRPCServer.GracefulStop()
	log.Info("application stopped successfully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
