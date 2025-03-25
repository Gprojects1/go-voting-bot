package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

func New(msg string) error {
	return customError{errorType: NoType, originalError: errors.New(msg)}
}

func Newf(msg string, args ...interface{}) error {
	return customError{errorType: NoType, originalError: errors.New(fmt.Sprintf(msg, args...))}
}

func Wrap(err error, msg string) error {
	return Wrapf(err, msg)
}

func Cause(err error) error {
	return errors.Cause(err)
}

func Wrapf(err error, msg string, args ...interface{}) error {
	wrappedError := errors.Wrapf(err, msg, args...)
	if customErr, ok := err.(customError); ok {
		return customError{
			errorType:     customErr.errorType,
			originalError: wrappedError,
			contextInfo:   customErr.contextInfo,
		}
	}

	return customError{errorType: NoType, originalError: wrappedError}
}
