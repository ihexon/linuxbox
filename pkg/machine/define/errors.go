package define

import (
	strongunits "bauklotze/pkg/storage"
	"errors"
	"fmt"
)

var (
	ErrNoSuchVM         = errors.New("VM does not exist")
	ErrWrongState       = errors.New("VM in wrong state to perform action")
	ErrVMAlreadyExists  = errors.New("VM already exists")
	ErrVMAlreadyRunning = errors.New("VM already running or starting")
	ErrMultipleActiveVM = errors.New("only one VM can be active at a time")
	ErrNotImplemented   = errors.New("functionality not implemented")
)

type ErrIncompatibleMachineConfig struct {
	Name string
	Path string
}

func (err *ErrIncompatibleMachineConfig) Error() string {
	return fmt.Sprintf("incompatible machine config %q (%s) for this version of Podman", err.Path, err.Name)
}

type ErrVMDoesNotExist struct {
	Name string
}

func (err *ErrVMDoesNotExist) Error() string {
	// the current error in qemu is not quoted
	return fmt.Sprintf("%s: VM does not exist", err.Name)
}

type ErrNewDiskSizeTooSmall struct {
	OldSize, NewSize strongunits.GiB
}

func (err *ErrNewDiskSizeTooSmall) Error() string {
	return fmt.Sprintf("invalid disk size %d: new disk must be larger than %dGB", err.OldSize, err.NewSize)
}
