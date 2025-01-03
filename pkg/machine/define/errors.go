//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

import (
	"errors"
	"fmt"
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
