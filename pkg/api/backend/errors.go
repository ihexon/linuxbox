//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package backend

import "errors"

var (
	ErrMachineConfigNull = errors.New("machineConfig is null")
	ErrStreamNotSupport  = errors.New("stream not support")
	ErrStopVMFailed      = errors.New("stop vm failed")
)
