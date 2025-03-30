package config

import (
	"os"
)

type Config struct {
	MattermostToken           string `json:"mattermost_token"`
	MattermostURL             string `json:"mattermost_url"`
	TarantoolHost             string `json:"tarantool"`
	TarantoolPort             string `json:"tarantool_port"`
	TarantoolUser             string `json:"tarantool_user"`
	TarantoolPassword         string `json:"tarantool_password"`
	AppPort                   string `json:"app_port"`
	Mattermost_url_web_socket string `json:"Mattermost_url_web_socket"`
}

func LoadConfig() (*Config, error) {
	config := Config{
		MattermostToken:           os.Getenv("BOT_TOKEN"),
		MattermostURL:             os.Getenv("MATTERMOST_URL"),
		TarantoolHost:             os.Getenv("TARANTOOL_HOST"),
		TarantoolPort:             os.Getenv("TARANTOOL_PORT"),
		TarantoolUser:             os.Getenv("TARANTOOL_USER"),
		TarantoolPassword:         os.Getenv("TARANTOOL_PASSWORD"),
		AppPort:                   os.Getenv("APP_PORT"),
		Mattermost_url_web_socket: os.Getenv("MATTERMOST_URL_WEB_SOCKET"),
	}

	// Проверка, что все необходимые переменные установлены (опционально)
	if config.MattermostToken == "" || config.MattermostURL == "" || config.TarantoolHost == "" || config.TarantoolPort == "" || config.TarantoolUser == "" || config.TarantoolPassword == "" || config.AppPort == "" || config.Mattermost_url_web_socket == "" {
		return nil, os.ErrNotExist // Или другая подходящая ошибка
	}

	return &config, nil
}
