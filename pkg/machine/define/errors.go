//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

import (
	"errors"
	"fmt"

	"github.com/containers/common/pkg/strongunits"
)

var (
	ErrVMAlreadyRunning = errors.New("VM already running or starting")
)

type IncompatibleMachineConfigError struct {
	Name string
	Path string
}

func (err *IncompatibleMachineConfigError) Error() string {
	return fmt.Sprintf("incompatible machine config %q (%s) for this version of Podman", err.Path, err.Name)
}

type VMDoesNotExistError struct {
	Name string
}

func (err *VMDoesNotExistError) Error() string {
	// the current error in qemu is not quoted
	return fmt.Sprintf("%s: VM does not exist", err.Name)
}

type NewDiskSizeTooSmallError struct {
	OldSize, NewSize strongunits.GiB
}

func (err *NewDiskSizeTooSmallError) Error() string {
	return fmt.Sprintf("invalid disk size %d: new disk must be larger than %dGB", err.OldSize, err.NewSize)
}
