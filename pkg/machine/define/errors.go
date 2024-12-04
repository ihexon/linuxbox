//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

import (
	"errors"
	"fmt"

	strongunits "github.com/containers/common/pkg/strongunits"
)

var (
	ErrWrongState       = errors.New("VM in wrong state to perform action")
	ErrVMAlreadyRunning = errors.New("VM already running or starting")
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
