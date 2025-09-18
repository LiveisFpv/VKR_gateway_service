package main

import (
	"VKR_gateway_service/internal/app"
	"VKR_gateway_service/internal/config"
	"VKR_gateway_service/internal/repository/postgres"
	"VKR_gateway_service/internal/transport/http"
	"VKR_gateway_service/pkg/logger"
	"VKR_gateway_service/pkg/storage"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title ALib API
// @version 0.1
// @description ALib для поиска различной научной информации
// @host localhost:8080
// @BasePath /api
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// ! Init logger
	logger := logger.LoggerSetup(true)
	// ! Parse config from env
	cfg, err := config.MustLoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load config with error: %v", err)
		return
	}
	// ! Init repoisitory
	// ! Init postgres
	pgPool, err := storage.PostgresConnect(ctx, cfg.PostgresConfig)
	if err != nil {
		logger.Fatalf("Failed to create pool conection to postgres with error: %v", err)
		return
	}

	UserRepo := postgres.NewUserRepository(pgPool)

	usecase := app.NewApp(cfg, UserRepo, logger)
	// ! Init REST
	// ! Init gRPC
	// ! Graceful shutdown
	server := http.NewHTTPServer(cfg, usecase)
	logger.Info("Start HTTP server")
	go func() {
		err = server.Listen()
		if err != nil {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	//Wait for interrupt signal to shutdown server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown HTTP Server ...")
	err = server.Stop(ctx)
	if err != nil {
		logger.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		logger.Info("Timeout stop server")
	default:
		logger.Info("Server exiting")
	}
}
