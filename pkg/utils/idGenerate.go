package utils

import "github.com/google/uuid"

func GenerateVotingID() string {
	return uuid.New().String()
}
