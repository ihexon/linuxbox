package define

import (
	"errors"
)

var (
	ErrVMAlreadyRunning = errors.New("VM already running or starting")
	ErrConstructVMFile  = errors.New("construct VMFile failed")
	ErrCatchSignal      = errors.New("catch signal")
	ErrPPIDNotRunning   = errors.New("PPID exited")
)
