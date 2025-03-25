package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	MattermostToken   string `json:"mattermost_token"`
	MattermostURL     string `json:"mattermost_url"`
	TarantoolHost     string `json:"tarantool"`
	TarantoolPort     string `json:"tarantool_port"`
	TarantoolUser     string `json:"tarantool_user"`
	TarantoolPassword string `json:"tarantool_password"`
	AppPort           string `json:"app_port"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
