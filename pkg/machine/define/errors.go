//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

import (
	"errors"
)

var (
	ErrVMAlreadyRunning = errors.New("VM already running or starting")
	ErrConstructVMFile  = errors.New("construct VMFile failed")
	ErrCatchSignal      = errors.New("catch signal")
	ErrPPIDNotRunning   = errors.New("PPID exited")
	ErrVMMExitNormally  = errors.New("hypervisor exited normally")
)
