package utils

import (
	"errors"
	"fmt"

	"github.com/containers/storage/pkg/regexp"
)

var (
	NameRegex     = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")
	ErrRegex      = fmt.Errorf("string must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidStr)
	ErrInvalidStr = errors.New("invalid strings")
)

func RegexValidation(s string) error {
	if !NameRegex.MatchString(s) {
		return fmt.Errorf("invalid name %q: %w", s, ErrRegex)
	}
	return nil
}
