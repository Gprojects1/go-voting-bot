package utils

import (
	"errors"
	"strconv"
)

func ParseInt(s string) (int, error) {
	if s == "" {
		return 0, errors.New("пустая строка")
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.New("не удалось преобразовать в целое число")
	}
	return value, nil
}
