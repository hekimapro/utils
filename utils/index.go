package utils

import "errors"

func CreateError(ErrorMessage string) error {
	return errors.New(ErrorMessage)
}