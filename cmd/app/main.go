package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/AndroDeMohawk/MovieVerse/internal/client/auth"
	"github.com/AndroDeMohawk/MovieVerse/internal/config"
	"github.com/AndroDeMohawk/MovieVerse/internal/infrastructure/kafka"
	"github.com/AndroDeMohawk/MovieVerse/internal/infrastructure/redis"
	"github.com/AndroDeMohawk/MovieVerse/internal/repository/postgres"
	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	commentgrpc "github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/comment"
	"github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/interceptor"
	moviegrpc "github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/movie"
	profilegrpc "github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/profile"
	httptransport "github.com/AndroDeMohawk/MovieVerse/internal/transport/http"
	"github.com/AndroDeMohawk/MovieVerse/internal/transport/ws"
	"github.com/AndroDeMohawk/MovieVerse/internal/usecase"
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

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	//1.PostgreSQL
	dbPool, err := postgres.New(appCtx, cfg.Postgres)
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()
	log.Info("connected to postgresql successfully")
	queries := db.New(dbPool)

	//2.Auth SSO Client
	authClient, err := auth.New(appCtx, cfg.AuthService.Address, cfg.AuthService.AppID, log)
	if err != nil {
		log.Error("failed to connect to auth service", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("connected to auth sso service successfully", slog.String("addr", cfg.AuthService.Address))

	//3.Redis Client
	pingCtx, pingCancel := context.WithTimeout(appCtx, 5*time.Second)
	defer pingCancel()

	rdb, err := redis.New(pingCtx, redis.Config{
		Addr:     cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})
	if err != nil {
		log.Error("failed to connect to redis", slog.Any("error", err))
		os.Exit(1)
	}

	//4.WebSocket Hub & Kafka
	wsHub := ws.NewHub(log)
	go wsHub.Run()

	kafkaProducer := kafka.NewProducer(cfg.Kafka.Brokers, log)
	defer kafkaProducer.Close()

	kafkaConsumer := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		"movie-comments-group",
		cfg.Kafka.CommentsTopic,
		wsHub,
		log,
	)
	kafkaConsumer.Start(appCtx)
	defer kafkaConsumer.Close()

	//5.Usecases
	movieUsecase := usecase.NewMovie(queries, dbPool, rdb, authClient, log)
	profileUsecase := usecase.NewProfile(queries, dbPool, log)
	commentUsecase := usecase.NewComment(
		queries,
		dbPool,
		kafkaProducer,
		cfg.Kafka.CommentsTopic,
		log,
	)

	go movieUsecase.StartViewsSync(appCtx, 1*time.Minute)

	//6.Transports Setup
	httpServer := httptransport.NewServer(strconv.Itoa(cfg.HTTP.Port), wsHub, log)

	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.AuthUnaryInterceptor(cfg.AuthService.AppSecret, cfg.AuthService.AppID),
		),
	)
	moviegrpc.Register(gRPCServer, movieUsecase)
	profilegrpc.Register(gRPCServer, profileUsecase)
	commentgrpc.Register(gRPCServer, commentUsecase)

	//7.Start HTTP Server
	go func() {
		if err := httpServer.Run(); err != nil {
			log.Error("HTTP server failed", slog.String("error", err.Error()))
		}
	}()

	//8.Start gRPC Server
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to listen port", slog.Int("port", cfg.GRPC.Port), slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		log.Info("gRPC server started", slog.String("addr", l.Addr().String()))
		if err := gRPCServer.Serve(l); err != nil && err != grpc.ErrServerStopped {
			log.Error("gRPC server failed", slog.String("error", err.Error()))
		}
	}()

	//9.Graceful Shutdown Wait
	<-appCtx.Done()
	log.Info("stopping application...")

	gRPCServer.GracefulStop()
	log.Info("gRPC server stopped")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.Error("failed to stop HTTP server", slog.String("error", err.Error()))
	} else {
		log.Info("HTTP server stopped gracefully")
	}

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
