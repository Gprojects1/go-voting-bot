package service

import (
	"go-voting-bot/pkg/dto"
	"go-voting-bot/pkg/errors"
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
		s.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали вопрос и как минимум два варианта ответа.  Пример: /poll create Вопрос | Вариант 1 | Вариант 2")
		err := errors.BadRequest.Wrapf(nil, errors.InvalidFormat.Message())
		err = errors.AddErrorContext(err, "message", "wrong question format, should be /poll create question | ans 1 | ans 2")
		return model.Voting{}, err
	}

	question := strings.TrimSpace(parts[0])
	options := make([]string, len(parts)-1)
	for i, option := range parts[1:] {
		options[i] = strings.TrimSpace(option)
	}

	if question == "" || len(options) < 2 {
		s.PostEphemeralMessage(channelID, userID, "Необходимо указать вопрос и как минимум два варианта ответа.")
		err := errors.BadRequest.Wrapf(nil, errors.InvalidFormat.Message())
		err = errors.AddErrorContext(err, "message", "wrong question format, should be /poll create question | ans 1 | ans 2 ...")
		return model.Voting{}, err
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
		s.PostEphemeralMessage(channelID, userID, "Неверный формат запроса.  Убедитесь, что вы указали /poll vote <id голосования> <номер варианта>")
		err := errors.BadRequest.Wrapf(nil, errors.InvalidFormat.Message())
		err = errors.AddErrorContext(err, "message", "wrong question format, should be /poll <vote voting id> <answer variant> ")
		return model.Voting{}, err
	}

	votingID := strings.TrimSpace(parts[0])
	optionNumberStr := strings.TrimSpace(parts[1])

	optionNumber, err := utils.ParseInt(optionNumberStr)
	if err != nil {
		s.PostEphemeralMessage(channelID, userID, "Номер варианта должен быть числом.")
		err = errors.BadRequest.Wrapf(err, errors.WrongType.Message())
		err = errors.AddErrorContext(err, "id", "wrong id format, should be an integer")
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
		err = errors.BadRequest.Wrapf(err, errors.UnavailableResource.Message())
		err = errors.AddErrorContext(err, "id", "Voting is finished")
		return model.Voting{}, err
	}

	if optionNumber < 1 || optionNumber > len(voting.Options) {
		s.PostEphemeralMessage(channelID, userID, "Неверный номер варианта.")
		err = errors.BadRequest.Wrapf(err, errors.BadRequest.Message())
		err = errors.AddErrorContext(err, "answer", "Wrong answer variant")
		return model.Voting{}, err
	}

	voting.Results[optionNumber-1]++
	s.Logger.Info("Vote registered", slog.String("voting_id", votingID), slog.Int("option_number", optionNumber), slog.String("user_id", userID))
	return s.VoteRepo.SaveVoting(voting)
}

func (s *VotingService) GetResultsByVotingId(request dto.VotingRequest, channelID string, userID string) (dto.VotingResultsResponse, string, error) {
	votingID := strings.TrimSpace(request.Text)
	if votingID == "" {
		s.Logger.Error("Wrong question format, should be /poll results <voting id>", slog.String("text", request.Text))
		s.PostEphemeralMessage(channelID, userID, "Используйте: /poll results <id голосования>")
		err := errors.BadRequest.Wrapf(nil, errors.InvalidFormat.Message())
		err = errors.AddErrorContext(err, "message", "wrong question format, should be /poll results <voting id>")
		return dto.VotingResultsResponse{}, "", err
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
		s.Logger.Error("Wrong question format, should be /poll end <voting id>", slog.String("text", request.Text))
		s.PostEphemeralMessage(channelID, userID, "Использование: /poll end <id голосования>")
		err := errors.BadRequest.Wrapf(nil, errors.InvalidFormat.Message())
		err = errors.AddErrorContext(err, "message", "wrong question format, should be /poll end <voting id>")
		return "", err
	}

	voting, err := s.VoteRepo.GetVoting(votingID)
	if err != nil {
		s.Logger.Error("Error getting voting from Tarantool" + err.Error())
		s.PostEphemeralMessage(channelID, userID, "Голосование не найдено.")
		return "", nil
	}

	if voting.CreatorID != userID {
		s.PostEphemeralMessage(channelID, userID, "Вы не являетесь создателем этого голосования.")
		err := errors.BadRequest.Wrapf(nil, errors.UnavailableResource.Message())
		err = errors.AddErrorContext(err, "id", "You are not a creator of this voting")
		return "", err
	}

	voting.IsActive = false
	voting.ClosedAt = time.Now()
	_, err = s.VoteRepo.SaveVoting(voting)

	s.Logger.Info("Voting ended", slog.String("voting_id", votingID), slog.String("user_id", userID))
	return votingID, err
}

func (s *VotingService) DeleteVotingByVotingId(request dto.VotingRequest, channelID, userID string) (string, error) {
	votingID := strings.TrimSpace(request.Text)
	if votingID == "" {
		s.Logger.Error("Wrong question format, should be /poll delete <voting id>", slog.String("text", request.Text))
		err := errors.BadRequest.Wrapf(nil, errors.InvalidFormat.Message())
		err = errors.AddErrorContext(err, "message", "wrong question format, should be /poll delete <voting id>")
		s.PostEphemeralMessage(channelID, userID, "Используйте: /poll delete <id голосования>")
		return "", err
	}

	voting, err := s.VoteRepo.GetVoting(votingID)
	if err != nil {
		s.Logger.Error("Error getting voting from Tarantool" + err.Error())
		s.PostEphemeralMessage(channelID, userID, "Голосование не найдено.")
		return "", err
	}

	if voting.CreatorID != userID {
		s.PostEphemeralMessage(channelID, userID, "Вы не являетесь создателем этого голосования.")
		err := errors.BadRequest.Wrapf(nil, errors.UnavailableResource.Message())
		err = errors.AddErrorContext(err, "id", "You are not a creator of this voting")
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
