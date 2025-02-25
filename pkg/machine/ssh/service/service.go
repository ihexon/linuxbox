//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package service

import (
	"context"
	"fmt"
	"time"

	"bauklotze/pkg/machine/vmconfig"
)

func GetKernelInfo(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return runCtx(ctx, mc, "uname", []string{
		"-a",
	})
}

func DoTimeSync(ctx context.Context, mc *vmconfig.MachineConfig) error {
	return runCtx(ctx, mc, "date", []string{
		"-s",
		fmt.Sprintf("@%d", time.Now().Unix()),
	})
}

func DoSync(mc *vmconfig.MachineConfig) error {
	return run(mc, "bash", []string{
		"-c",
		"sync",
	})
}
