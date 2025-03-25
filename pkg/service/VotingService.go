package service

import (
	"go-voting-bot/pkg/dto"
	"go-voting-bot/pkg/model"
	"go-voting-bot/pkg/repository"
	"go-voting-bot/pkg/utils"
	"log/slog"
	"strings"
	"time"

	mattermodel "github.com/mattermost/mattermost-server/v6/model"
)

type VotingService struct {
	Client   *mattermodel.Client4
	VoteRepo repository.VotingRepository
	Logger   *slog.Logger
}

func (s *VotingService) AddNewVoting(request dto.VotingRequest, channelID, userID string) (model.Voting, error) {
	parts := strings.Split(request.Text, "|")
	if len(parts) < 3 {
		s.Logger.Error("Invalid format: requires question and at least two options", slog.String("text", request.Text))
		s.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали вопрос и как минимум два варианта ответа.  Пример: /create Вопрос | Вариант 1 | Вариант 2")

		return model.Voting{}, nil
	}

	question := strings.TrimSpace(parts[0])
	options := make([]string, len(parts)-1)
	for i, option := range parts[1:] {
		options[i] = strings.TrimSpace(option)
	}

	if question == "" || len(options) < 2 {
		s.PostEphemeralMessage(channelID, userID, "Необходимо указать вопрос и как минимум два варианта ответа.")
		return model.Voting{}, nil
	}

	voting := model.Voting{
		ID:        utils.GenerateVotingID(),
		CreatorID: userID,
		ChannelID: channelID,
		Question:  question,
		Options:   options,
		CreatedAt: time.Now(),
		Results:   make(map[int]int),
		IsActive:  true,
	}
	return s.VoteRepo.SaveVoting(voting)
}

func (s *VotingService) AddNewVote(request dto.VotingRequest, channelID, userID string) (model.Voting, error) {
	parts := strings.Split(request.Text, " ")
	if len(parts) != 2 {
		s.Logger.Error("Invalid format: requires voting id and 1 answer", slog.String("text", request.Text))
		s.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали вопрос и как минимум два варианта ответа.Пример: /vote <id голосования> <номер варианта>")
		return model.Voting{}, nil
	}

	votingID := strings.TrimSpace(parts[0])
	optionNumberStr := strings.TrimSpace(parts[1])

	optionNumber, err := utils.ParseInt(optionNumberStr)
	if err != nil {
		s.PostEphemeralMessage(channelID, userID, "Номер варианта должен быть числом.")
		return model.Voting{}, err
	}

	voting, err := s.VoteRepo.GetVoting(votingID)
	if err != nil {
		s.Logger.Error("Error getting voting from Tarantool" + err.Error())
		s.PostEphemeralMessage(channelID, userID, "Голосование не найдено.")
		return model.Voting{}, err
	}

	if !voting.IsActive {
		s.PostEphemeralMessage(channelID, userID, "Голосование завершено и больше не принимает голоса.")
		return model.Voting{}, err
	}

	if optionNumber < 1 || optionNumber > len(voting.Options) {
		s.PostEphemeralMessage(channelID, userID, "Неверный номер варианта.")
		return model.Voting{}, err
	}

	voting.Results[optionNumber-1]++
	s.Logger.Info("Vote registered", slog.String("voting_id", votingID), slog.Int("option_number", optionNumber), slog.String("user_id", userID))
	return s.VoteRepo.SaveVoting(voting)
}

func (s *VotingService) GetResultsByVotingId(request dto.VotingRequest, channelID string, userID string) (dto.VotingResultsResponse, string, error) {
	votingID := strings.TrimSpace(request.Text)
	if votingID == "" {
		s.PostEphemeralMessage(channelID, userID, "Используйте: /results <id голосования>")
		return dto.VotingResultsResponse{}, "", nil
	}
	voting, err := s.VoteRepo.GetVoting(votingID)
	if err != nil {
		return dto.VotingResultsResponse{}, "", err
	}

	totalVotes := 0
	results := make([]dto.Result, len(voting.Options))

	for i, option := range voting.Options {
		votes := voting.Results[i]
		totalVotes += votes
		percentage := 0.0
		if totalVotes > 0 {
			percentage = float64(votes) / float64(totalVotes) * 100
		}
		results[i] = dto.Result{
			Option:     option,
			VoteCount:  votes,
			Percentage: percentage,
		}
	}

	return dto.VotingResultsResponse{
		Question:   voting.Question,
		Options:    voting.Options,
		Results:    results,
		TotalVotes: totalVotes,
	}, votingID, nil
}

func (s *VotingService) EndVotingByVotingId(request dto.VotingRequest, channelID, userID string) (string, error) {
	votingID := strings.TrimSpace(request.Text)
	if votingID == "" {
		s.PostEphemeralMessage(channelID, userID, "Использование: /end <id голосования>")
		return "", nil
	}

	voting, err := s.VoteRepo.GetVoting(votingID)
	if err != nil {
		s.Logger.Error("Error getting voting from Tarantool" + err.Error())
		s.PostEphemeralMessage(channelID, userID, "Голосование не найдено.")
		return "", nil
	}

	if voting.CreatorID != userID {
		s.PostEphemeralMessage(channelID, userID, "Вы не являетесь создателем этого голосования.")
		return "", nil
	}

	voting.IsActive = false
	voting.ClosedAt = time.Now()
	//delete voting
	_, err = s.VoteRepo.SaveVoting(voting)

	s.Logger.Info("Voting ended", slog.String("voting_id", votingID), slog.String("user_id", userID))
	return votingID, err
}

func (s *VotingService) DeleteVotingByVotingId(request dto.VotingRequest, channelID, userID string) (string, error) {
	votingID := strings.TrimSpace(request.Text)
	if votingID == "" {
		s.PostEphemeralMessage(channelID, userID, "Используйте: /delete <id голосования>")
		return "", nil
	}

	voting, err := s.VoteRepo.GetVoting(votingID)
	if err != nil {
		s.Logger.Error("Error getting voting from Tarantool" + err.Error())
		s.PostEphemeralMessage(channelID, userID, "Голосование не найдено.")
		return "", err
	}

	if voting.CreatorID != userID {
		s.PostEphemeralMessage(channelID, userID, "Вы не являетесь создателем этого голосования.")
		return "", err
	}
	s.Logger.Info("Voting deleted", slog.String("voting_id", votingID), slog.String("user_id", userID))

	return s.VoteRepo.DeleteVoting(votingID)
}

func (s *VotingService) PostMessage(channelID, message string) {
	post := &mattermodel.Post{
		ChannelId: channelID,
		Message:   message,
	}
	_, _, err := s.Client.CreatePost(post)
	if err != nil {
		s.Logger.Error("Failed to post message to Mattermost", slog.String("channel_id", channelID), slog.Any("error", err))
	}
}

func (s *VotingService) PostEphemeralMessage(channelID, userID, message string) {
	post := &mattermodel.PostEphemeral{
		UserID: userID,
		Post: &mattermodel.Post{
			Message:   message,
			ChannelId: channelID,
		},
	}

	_, _, err := s.Client.CreatePostEphemeral(post)
	if err != nil {
		s.Logger.Error("Failed to post ephemeral message to Mattermost",
			slog.String("channel_id", channelID),
			slog.String("user_id", userID),
			slog.Any("error", err))
	}
}
