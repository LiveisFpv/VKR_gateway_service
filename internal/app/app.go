package app

import (
	"VKR_gateway_service/internal/config"
	"VKR_gateway_service/internal/repository"

	"github.com/sirupsen/logrus"
)

type App struct {
	Config *config.Config
	Logger *logrus.Logger
}

func NewApp(
	cfg *config.Config,
	UserRepository repository.UserRepository,
	Logger *logrus.Logger,
) *App {
	return &App{
		Config: cfg,
		Logger: Logger,
	}
}
