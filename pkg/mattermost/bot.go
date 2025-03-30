package mattermost

import (
	"encoding/json"
	"fmt"
	"go-voting-bot/config"
	"go-voting-bot/pkg/controller"
	"go-voting-bot/pkg/dto"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-server/v6/model"
)

type MattermostBot struct {
	BotID      string
	TeamID     string
	Controller *controller.VotingController
	Logger     *slog.Logger
	ServerURL  string
	Token      string
	AppPort    string
	Ws_URL     string
}

func NewMattermostBot(cfg *config.Config, con *controller.VotingController, logger *slog.Logger) (*MattermostBot, error) {
	con.Service.Client.SetToken(cfg.MattermostToken)

	user, _, err := con.Service.Client.GetUser("me", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get bot user: %w", err)
	}

	teams, _, err := con.Service.Client.GetTeamsForUser(user.Id, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get teams for user: %w", err)
	}

	if len(teams) == 0 {
		return nil, fmt.Errorf("bot is not a member of any team")
	}

	return &MattermostBot{
		BotID:      user.Id,
		TeamID:     teams[0].Id,
		Controller: con,
		Logger:     logger,
		ServerURL:  cfg.MattermostURL,
		Token:      cfg.MattermostToken,
		AppPort:    cfg.AppPort,
		Ws_URL:     cfg.Mattermost_url_web_socket,
	}, nil
}

func (b *MattermostBot) Start() {
	b.Logger.Info("Mattermost bot started")

	go b.listenToEvents()

	router := gin.Default()
	port := b.AppPort
	address := ":" + port
	b.Logger.Info("Starting HTTP server with Gin", slog.String("port", port))
	if err := router.Run(address); err != nil {
		b.Logger.Error("Failed to start HTTP server with Gin" + err.Error())
	}
}

func (b *MattermostBot) listenToEvents() {
	failCount := 0

	for {
		b.Logger.Info("Connecting to Mattermost WebSocket")
		fmt.Println(b.Token)
		fmt.Println(b.Ws_URL)
		webSocketClient, err := model.NewWebSocketClient4(b.Ws_URL, b.Token)
		if err != nil {
			b.Logger.Warn("Failed to connect to WebSocket, retrying...", slog.Any("error", err))
			failCount++
			time.Sleep(time.Duration(failCount) * time.Second)
			continue
		}
		b.Logger.Info("Connected to Mattermost WebSocket")

		webSocketClient.Listen()

		for event := range webSocketClient.EventChannel {
			go b.handleWebSocketEvent(event)
		}
	}
}

func (b *MattermostBot) handleWebSocketEvent(event *model.WebSocketEvent) {
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), post)
	if err != nil {
		//b.Logger.Error().Err(err).Msg("Could not unmarshal post from WebSocket event")
		return
	}

	if post.UserId == b.BotID {
		return
	}
	c, _ := gin.CreateTestContext(nil)
	b.processCommand(c, post)
}

func (b *MattermostBot) processCommand(c *gin.Context, post *model.Post) {
	if !strings.HasPrefix(post.Message, "/poll") {
		return
	}

	args := strings.Fields(post.Message)
	if len(args) < 2 {
		b.Controller.Service.PostMessage(post.ChannelId, "Ипспользуйте:/poll create, vote, results, close, delete")
		return
	}

	action := strings.ToLower(args[1])

	messageBody := strings.Join(args[2:], " ")
	messageBody = strings.TrimSpace(messageBody)

	dto := dto.CommandRequest{
		Message:   messageBody,
		UserID:    post.UserId,
		ChannelID: post.ChannelId,
	}

	switch action {
	case "create":
		b.Controller.CreateVoting(c, dto)
	case "vote":
		b.Controller.AddVote(c, dto)
	case "results":
		b.Controller.GetResults(c, dto)
	case "close":
		b.Controller.EndVoting(c, dto)
	case "delete":
		b.Controller.DeleteVoting(c, dto)
	default:
		b.Controller.Service.PostMessage(post.ChannelId, "Недопустимая команда. Ипспользуйте: create, vote, results, close, delete")
	}
}
func SetupGracefulShutdown(bot *MattermostBot) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		bot.Logger.Info("Shutting down...")
		os.Exit(0)
	}()
}
