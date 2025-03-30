package main

import (
	"go-voting-bot/config"
	"go-voting-bot/pkg/controller"
	"go-voting-bot/pkg/mattermost"
	"go-voting-bot/pkg/repository"
	"go-voting-bot/pkg/service"
	"log"
	"log/slog"
	"os"

	"github.com/mattermost/mattermost-server/v6/model"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	handler := slog.NewTextHandler(os.Stdout, nil)

	logger := slog.New(handler)

	votingRepo, err := repository.NewTarantoolClient(cfg.TarantoolHost, cfg.TarantoolPort, cfg.TarantoolUser, cfg.TarantoolPassword)
	if err != nil {
		logger.Error("Ошибка создания подключения к базе данных", slog.Any("error", err))
		return
	}
	votingService := &service.VotingService{
		Client:   model.NewAPIv4Client(cfg.MattermostURL),
		VoteRepo: votingRepo,
		Logger:   logger,
	}

	votingController := &controller.VotingController{
		Service: votingService,
		Logger:  logger,
	}

	mattermostBot, err := mattermost.NewMattermostBot(cfg, votingController, logger)
	if err != nil {
		logger.Error("Ошибка создания Mattermost бота", slog.Any("error", err))
		return
	}

	mattermostBot.Start()
	mattermost.SetupGracefulShutdown(mattermostBot)
}
