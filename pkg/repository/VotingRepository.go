package repository

import (
	"fmt"
	"go-voting-bot/pkg/model"
	"go-voting-bot/pkg/utils"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/tarantool/go-tarantool"
)

type VotingRepository interface {
	SaveVoting(voting model.Voting) (model.Voting, error)
	GetVoting(votingID string) (model.Voting, error)
	DeleteVoting(votingID string) (string, error)
}

type votingRepository struct {
	Conn *tarantool.Connection
}

func NewTarantoolClient(host string, port string, user string, password string) (*votingRepository, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := tarantool.Connect(addr, tarantool.Opts{
		User: user,
		Pass: password,
	})
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
		return model.Voting{}, fmt.Errorf("failed to save voting: %w", err)
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
		return model.Voting{}, fmt.Errorf("failed to get voting from Tarantool: %w", err)
	}

	if len(resp.Data) == 0 {
		return model.Voting{}, fmt.Errorf("voting not found for ID: %s", votingID)
	}

	if err := mapstructure.Decode(resp.Data[0], &result); err != nil {
		return model.Voting{}, fmt.Errorf("failed to decode voting data: %w", err)
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
		return "", fmt.Errorf("failed to delete voting: %w", err)
	}
	return votingID, nil
}
