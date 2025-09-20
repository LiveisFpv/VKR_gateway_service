package app

import (
    "VKR_gateway_service/internal/config"
    "VKR_gateway_service/internal/repository"
    pb "VKR_gateway_service/gen/go"

    "github.com/sirupsen/logrus"
)

type App struct {
    Config *config.Config
    Logger *logrus.Logger
    // gRPC client for external AI service
    AI     pb.SemanticServiceClient
}

func NewApp(
    cfg *config.Config,
    UserRepository repository.UserRepository,
    Logger *logrus.Logger,
    AI pb.SemanticServiceClient,
) *App {
    return &App{
        Config: cfg,
        Logger: Logger,
        AI:     AI,
    }
}
