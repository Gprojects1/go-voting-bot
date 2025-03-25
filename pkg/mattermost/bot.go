package mattermost

import (
	"fmt"
	"go-voting-bot/config"
	"go-voting-bot/pkg/controller"
	"log/slog"
	"net/http"

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
}

func NewMattermostBot(cfg *config.Config, con *controller.VotingController, logger *slog.Logger) (*MattermostBot, error) {
	con.Service.Client.SetToken(cfg.MattermostToken)

	user, _, err := con.Service.Client.GetMe("")
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
	}, nil
}

func (b *MattermostBot) Start() {
	b.Logger.Info("Mattermost bot started")

	err := b.RegisterCommands()
	if err != nil {
		b.Logger.Error("Failed to register commands" + err.Error())
	}

	router := gin.Default()
	router.POST("/", b.handleMattermostRequests)

	port := b.AppPort
	b.Logger.Info("Starting HTTP server with Gin", slog.String("port", port))
	err = router.Run(port)
	if err != nil {
		b.Logger.Error("Failed to start HTTP server with Gin" + err.Error())
	}
}

func (b *MattermostBot) handleMattermostRequests(c *gin.Context) {
	b.Logger.Info("Received POST request from Mattermost")

	command := c.PostForm("command")
	channelID := c.PostForm("channel_id")
	userID := c.PostForm("user_id")
	token := c.PostForm("token")

	b.Logger.Info("Request details",
		slog.String("command", command),
		slog.String("channel_id", channelID),
		slog.String("user_id", userID),
	)

	if token != "43z6gcape3fcxrswhp7sdxjn4o" {
		b.Logger.Warn("Invalid token received")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	switch command {
	case "/create":
		b.Controller.CreateVoting(c)
	case "/vote":
		b.Controller.AddVote(c)
	case "/results":
		b.Controller.GetResults(c)
	case "/end":
		b.Controller.EndVoting(c)
	case "/delete":
		b.Controller.DeleteVoting(c)
	default:
		b.Logger.Warn("Unknown command received", slog.String("command", command))
		c.AbortWithStatus(http.StatusBadRequest)
	}
}

func (b *MattermostBot) RegisterCommands() error {
	commands := []*model.Command{
		{
			Trigger:          "create",
			Method:           "POST",
			URL:              b.ServerURL,
			AutoComplete:     true,
			AutoCompleteDesc: "Create a new poll",
			AutoCompleteHint: "<question> | <option1> | <option2> | ...",
			DisplayName:      "Create Poll",
			Description:      "Create a new poll in the channel",
			Token:            b.Token,
		},
		{
			Trigger:          "vote",
			Method:           "POST",
			URL:              b.ServerURL,
			AutoComplete:     true,
			AutoCompleteDesc: "Vote in a poll",
			AutoCompleteHint: "<poll_id> <option_number>",
			DisplayName:      "Vote",
			Description:      "Vote for an option in a poll",
			Token:            b.Token,
		},
		{
			Trigger:          "results",
			Method:           "POST",
			URL:              b.ServerURL,
			AutoComplete:     true,
			AutoCompleteDesc: "View poll results",
			AutoCompleteHint: "<poll_id>",
			DisplayName:      "Poll Results",
			Description:      "View the results of a poll",
			Token:            b.Token,
		},
		{
			Trigger:          "end",
			Method:           "POST",
			URL:              b.ServerURL,
			AutoComplete:     true,
			AutoCompleteDesc: "End a poll",
			AutoCompleteHint: "<poll_id>",
			DisplayName:      "End Poll",
			Description:      "End a poll and prevent further voting",
			Token:            b.Token,
		},
		{
			Trigger:          "delete",
			Method:           "POST",
			URL:              b.ServerURL,
			AutoComplete:     true,
			AutoCompleteDesc: "Delete a poll",
			AutoCompleteHint: "<poll_id>",
			DisplayName:      "Delete Poll",
			Description:      "Delete a poll (creator only)",
			Token:            b.Token,
		},
	}

	for _, command := range commands {
		_, _, err := b.Controller.Service.Client.CreateCommand(command)
		if err != nil {
			b.Logger.Error("Failed to register command", slog.String("command", command.Trigger), slog.Any("error", err))
			return fmt.Errorf("failed to register command %s: %w", command.Trigger, err)
		}
		b.Logger.Info("Registered command", slog.String("command", command.Trigger))
	}
	return nil
}
