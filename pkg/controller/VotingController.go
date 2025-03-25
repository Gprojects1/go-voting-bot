package controller

import (
	"fmt"
	"go-voting-bot/pkg/dto"
	"go-voting-bot/pkg/errors"
	"go-voting-bot/pkg/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VotingController struct {
	Service *service.VotingService
	Logger  *slog.Logger
}

func (con *VotingController) CreateVoting(c *gin.Context) {
	channelID := c.PostForm("channel_id")
	userID := c.PostForm("user_id")

	var request dto.VotingRequest
	if err := c.ShouldBind(&request); err != nil {
		con.Logger.Error("Error binding request:" + err.Error())
		con.Service.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали вопрос и как минимум два варианта ответа.  Пример: /create Вопрос | Вариант 1 | Вариант 2")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	voting, err := con.Service.AddNewVoting(request, channelID, userID)
	if err != nil {
		con.Service.PostEphemeralMessage(channelID, userID, "Произошла ошибка при создании голосования.")
		errors.ErrorHandler(c, err)
	}
	message := fmt.Sprintf("Голосование создано!\n**%s**\n", voting.Question)
	for i, option := range voting.Options {
		message += fmt.Sprintf(":white_check_mark: %s - `/vote %s %d`\n", option, voting.ID, i+1)
	}
	message += fmt.Sprintf("\nЧтобы просмотреть результаты, используйте `/results %s`", voting.ID)
	message += fmt.Sprintf("\nЧтобы завершить голосование, используйте `/end %s`", voting.ID)

	con.Service.PostMessage(channelID, message)

	con.Service.PostEphemeralMessage(channelID, userID, fmt.Sprintf("Голосование с ID `%s` создано.", voting.ID))

	c.Status(http.StatusOK)
}

func (con *VotingController) AddVote(c *gin.Context) {
	channelID := c.PostForm("channel_id")
	userID := c.PostForm("user_id")

	var request dto.VotingRequest
	if err := c.ShouldBind(&request); err != nil {
		con.Logger.Error("Error binding request:" + err.Error())
		con.Service.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали вопрос и как минимум два варианта ответа. Пример: /vote <id голосования> <номер варианта>")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	con.Logger.Info("Handling /vote command", slog.String("channel_id", channelID), slog.String("user_id", userID))

	_, err := con.Service.AddNewVote(request, channelID, userID)
	if err != nil {
		con.Service.PostEphemeralMessage(channelID, userID, "Произошла ошибка при обработке голоса.")
		errors.ErrorHandler(c, err)
	}

	message := fmt.Sprintf("Новый голос учтён!")

	con.Service.PostMessage(channelID, message)

	con.Service.PostEphemeralMessage(channelID, userID, fmt.Sprintf("Ваш голос учтён!"))

	c.Status(http.StatusOK)
}

func (con *VotingController) GetResults(c *gin.Context) {
	channelID := c.PostForm("channel_id")
	userID := c.PostForm("user_id")

	var request dto.VotingRequest
	if err := c.ShouldBind(&request); err != nil {
		con.Logger.Error("Error binding request:" + err.Error())
		con.Service.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали вопрос и как минимум два варианта ответа. Пример: /results <id голосования>")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	con.Logger.Info("Handling /results command", slog.String("channel_id", channelID), slog.String("user_id", userID))

	votingResults, VotingID, err := con.Service.GetResultsByVotingId(request, channelID, userID)
	if err != nil {
		con.Service.PostEphemeralMessage(channelID, userID, "Произошла ошибка при обработке голоса.")
		errors.ErrorHandler(c, err)
		return
	}

	message := fmt.Sprintf("**Результаты голосования: %s**\n", votingResults.Question)

	for i, result := range votingResults.Results {
		message += fmt.Sprintf("%d. %s: %d (%.2f%%)\n", i+1, result.Option, result.VoteCount, result.Percentage)
	}

	con.Logger.Info("Results requested", slog.String("voting_id", VotingID), slog.String("user_id", userID))

	con.Service.PostMessage(channelID, message)

	c.Status(http.StatusOK)
}

func (con *VotingController) EndVoting(c *gin.Context) {
	channelID := c.PostForm("channel_id")
	userID := c.PostForm("user_id")

	var request dto.VotingRequest
	if err := c.ShouldBind(&request); err != nil {
		con.Logger.Error("Error binding request:" + err.Error())
		con.Service.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали id голосования. Пример: /end <id голосования>")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	con.Logger.Info("Handling /end command", slog.String("channel_id", channelID), slog.String("user_id", userID))

	votingID, err := con.Service.EndVotingByVotingId(request, channelID, userID)
	if err != nil {
		con.Service.PostEphemeralMessage(channelID, userID, "Произошла ошибка при обработке голоса.")
		errors.ErrorHandler(c, err)
		return
	}

	message := fmt.Sprintf("Голосование **%s** завершено.", votingID)
	con.Service.PostMessage(channelID, message)

	c.Status(http.StatusOK)
}

func (con *VotingController) DeleteVoting(c *gin.Context) {
	channelID := c.PostForm("channel_id")
	userID := c.PostForm("user_id")

	var request dto.VotingRequest
	if err := c.ShouldBind(&request); err != nil {
		con.Logger.Error("Error binding request:" + err.Error())
		con.Service.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали id голосования. Пример: /delete <id голосования>")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	con.Logger.Info("Handling /delete command", slog.String("channel_id", channelID), slog.String("user_id", userID))

	votingID, err := con.Service.DeleteVotingByVotingId(request, channelID, userID)
	if err != nil {
		con.Service.PostEphemeralMessage(channelID, userID, "Произошла ошибка при обработке голоса.")
		errors.ErrorHandler(c, err)
		return
	}

	message := fmt.Sprintf("Голосование **%s** удалено.", votingID)
	con.Service.PostEphemeralMessage(channelID, userID, message)
	c.Status(http.StatusOK)
}
