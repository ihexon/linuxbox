package define

import (
	"errors"
	"fmt"

	"github.com/containers/storage/pkg/regexp"
)

var (
	NameRegex     = regexp.Delayed("^[a-zA-Z0-9][a-zA-Z0-9_.-]*$")
	ErrRegex      = fmt.Errorf("names must match [a-zA-Z0-9][a-zA-Z0-9_.-]*: %w", ErrInvalidArg)
	ErrInvalidArg = errors.New("invalid argument")
)
