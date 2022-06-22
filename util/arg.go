package util

import (
	"errors"
)

func ValidateStringArgs(args []string) error {
	if len(args) > 0 && len(args[0]) > 0 {
		return nil
	}

	return errors.New("invalid config file")
}

func VarPtr[T any](arg T) *T {
	return &arg
}
