package repository

import (
	"fmt"
	"go-voting-bot/pkg/model"
	"go-voting-bot/pkg/utils"
	"time"

	"context"
	"go-voting-bot/pkg/errors"
	"log/slog"

	"github.com/mitchellh/mapstructure"
	"github.com/tarantool/go-tarantool/v2"
)

type VotingRepository interface {
	SaveVoting(voting model.Voting) (model.Voting, error)
	GetVoting(votingID string) (model.Voting, error)
	DeleteVoting(votingID string) (string, error)
}

type votingRepository struct {
	Conn   *tarantool.Connection
	Logger *slog.Logger
}

func NewTarantoolClient(host string, port string, user string, password string) (*votingRepository, error) {
	addr := fmt.Sprintf("%s:%s", host, port)

	dialer := tarantool.NetDialer{
		Address:  addr,
		User:     "guest",
		Password: "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := tarantool.Connect(ctx, dialer, tarantool.Opts{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Tarantool: %w", err)
	}

	return &votingRepository{Conn: conn}, nil
}

func (t *votingRepository) Close() error {
	return t.Conn.Close()
}

func (t *votingRepository) SaveVoting(voting model.Voting) (model.Voting, error) {
	_, err := t.Conn.Insert("votings", []interface{}{
		voting.ID,
		voting.CreatorID,
		voting.ChannelID,
		voting.Question,
		voting.Options,
		voting.CreatedAt,
		voting.ClosedAt,
		voting.IsActive,
		utils.ConvertResultsToMapStringInterface(voting.Results),
	})
	if err != nil {
		t.Logger.Error("can't save record with this id", slog.String("id", voting.ID))
		err = errors.Wrapf(err, errors.NotSaved.Message())
		err = errors.AddErrorContext(err, voting.ID, "can't save record with this id")
		return model.Voting{}, err
	}
	return voting, nil
}

func (t *votingRepository) GetVoting(votingID string) (model.Voting, error) {
	var result []struct {
		ID        string
		CreatorID string
		ChannelID string
		Question  string
		Options   []string
		CreatedAt time.Time
		ClosedAt  time.Time
		IsActive  bool
		Results   map[string]interface{}
	}

	resp, err := t.Conn.Select("votings", "primary", 0, 1, tarantool.IterEq, []interface{}{votingID})
	if err != nil {
		t.Logger.Error("Failed to get voting from Tarantool", slog.String("id", votingID))
		err = errors.Wrapf(err, errors.NotFound.Message())
		err = errors.AddErrorContext(err, votingID, "Failed to get voting from Tarantool")
		return model.Voting{}, err
	}

	if len(resp) == 0 {
		t.Logger.Error("Voting not found for ID", slog.String("id", votingID))
		err = errors.Wrapf(err, errors.WrongType.Message())
		err = errors.AddErrorContext(err, votingID, "Voting not found for ID")
		return model.Voting{}, err
	}

	if err := mapstructure.Decode(resp[0], &result); err != nil {
		t.Logger.Error("Failed to decode voting data", slog.String("id", votingID))
		err = errors.Wrapf(err, errors.InvalidFormat.Message())
		err = errors.AddErrorContext(err, votingID, "Failed to decode voting data")
		return model.Voting{}, err
	}

	voting := model.Voting{
		ID:        result[0].ID,
		CreatorID: result[0].CreatorID,
		ChannelID: result[0].ChannelID,
		Question:  result[0].Question,
		Options:   result[0].Options,
		CreatedAt: result[0].CreatedAt,
		ClosedAt:  result[0].ClosedAt,
		IsActive:  result[0].IsActive,
		Results:   utils.ConvertMapStringInterfaceToResults(result[0].Results),
	}

	return voting, nil
}

func (t *votingRepository) DeleteVoting(votingID string) (string, error) {
	_, err := t.Conn.Delete("votings", "primary", []interface{}{votingID})
	if err != nil {
		t.Logger.Error("Failed to delete voting data", slog.String("id", votingID))
		err = errors.Wrapf(err, errors.NotFound.Message())
		err = errors.AddErrorContext(err, votingID, "Failed to delete voting data")
		return "", err
	}
	return votingID, nil
}
